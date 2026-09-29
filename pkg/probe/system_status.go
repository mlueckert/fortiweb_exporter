// Copyright (C) 2025
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package probe

import (
	"log"
	"strconv"

	"github.com/mlueckert/fortiweb_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

type systemStatusClusterMember struct {
	Hostname string `json:"hostname"`
	DevSN    string `json:"dev_sn"`
	Role     string `json:"role"`
}

// systemStatusResult is the result of the FortiWeb monitor API endpoint
// api/v2.0/system/status.systemstatus.
type systemStatusResult struct {
	Cluster         string                      `json:"cluster"`
	ClusterMembers  []systemStatusClusterMember `json:"cluster_members"`
	SerialNumber    string                      `json:"serialNumber"`
	FirmwareVersion string                      `json:"firmwareVersion"`
	UpDays          string                      `json:"up_days"`
	UpHrs           string                      `json:"up_hrs"`
	UpMins          string                      `json:"up_mins"`
}

type systemStatusResponse struct {
	Results systemStatusResult `json:"results"`
}

func probeSystemStatus(c http.FortiHTTP) ([]prometheus.Metric, bool) {
	var (
		mInfo = prometheus.NewDesc(
			"fortiweb_system_info",
			"System information, as reported by system/status.systemstatus",
			[]string{"serial", "firmware_version", "cluster_name", "cluster_role"}, nil,
		)
		mUptime = prometheus.NewDesc(
			"fortiweb_uptime_seconds",
			"Time since the system was started in seconds (minute resolution), as reported by system/status.systemstatus",
			nil, nil,
		)
	)

	var r systemStatusResponse
	if err := c.Get("api/v2.0/system/status.systemstatus", "", &r); err != nil {
		log.Printf("Error: %v", err)
		return nil, false
	}
	res := r.Results

	role := ""
	for _, m := range res.ClusterMembers {
		if m.DevSN == res.SerialNumber {
			role = m.Role
			break
		}
	}

	m := []prometheus.Metric{
		prometheus.MustNewConstMetric(mInfo, prometheus.GaugeValue, 1, res.SerialNumber, res.FirmwareVersion, res.Cluster, role),
	}

	uptime := 0
	for _, part := range []struct {
		v    string
		mult int
	}{{res.UpDays, 86400}, {res.UpHrs, 3600}, {res.UpMins, 60}} {
		n, err := strconv.Atoi(part.v)
		if err != nil {
			log.Printf("Error: unable to parse uptime from system/status.systemstatus: %v", err)
			return m, false
		}
		uptime += n * part.mult
	}
	m = append(m, prometheus.MustNewConstMetric(mUptime, prometheus.GaugeValue, float64(uptime)))

	return m, true
}
