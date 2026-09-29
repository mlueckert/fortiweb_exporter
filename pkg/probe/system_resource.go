// Copyright (C) 2025
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package probe

import (
	"log"

	"github.com/mlueckert/fortiweb_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

// systemResourceResult mirrors the "results" object returned by the FortiWeb
// monitor API endpoint api/v2.0/system/status.systemresource.
type systemResourceResult struct {
	CPU           int    `json:"cpu"`
	Mem           int    `json:"mem"`
	LogDisk       string `json:"logDisk"`
	DBStatus      string `json:"dbStatus"`
	DiskUsage     int    `json:"diskUsage"`
	SessionCount  int    `json:"sessionCount"`
	ConnCntPerSec int    `json:"connCntPerSec"`
}

type systemResourceResponse struct {
	Results systemResourceResult `json:"results"`
}

func probeSystemResource(c http.FortiHTTP) ([]prometheus.Metric, bool) {
	var (
		mCPU = prometheus.NewDesc(
			"fortiweb_cpu_usage_ratio",
			"Current CPU usage ratio, as reported by system/status.systemresource",
			nil, nil,
		)
		mMem = prometheus.NewDesc(
			"fortiweb_memory_usage_ratio",
			"Current memory usage ratio, as reported by system/status.systemresource",
			nil, nil,
		)
		mDiskUsage = prometheus.NewDesc(
			"fortiweb_disk_usage_ratio",
			"Current disk usage ratio, as reported by system/status.systemresource",
			nil, nil,
		)
		mSessionCount = prometheus.NewDesc(
			"fortiweb_current_sessions",
			"Current amount of sessions, as reported by system/status.systemresource",
			nil, nil,
		)
		mConnCntPerSec = prometheus.NewDesc(
			"fortiweb_connections_per_second",
			"Current amount of new connections per second, as reported by system/status.systemresource",
			nil, nil,
		)
		mLogDisk = prometheus.NewDesc(
			"fortiweb_log_disk_info",
			"Status of the log disk, as reported by system/status.systemresource",
			[]string{"status"}, nil,
		)
		mDBStatus = prometheus.NewDesc(
			"fortiweb_db_status_info",
			"Status of the local database, as reported by system/status.systemresource",
			[]string{"status"}, nil,
		)
	)

	var r systemResourceResponse
	if err := c.Get("api/v2.0/system/status.systemresource", "", &r); err != nil {
		log.Printf("Error: %v", err)
		return nil, false
	}

	m := []prometheus.Metric{
		prometheus.MustNewConstMetric(mCPU, prometheus.GaugeValue, float64(r.Results.CPU)/100.0),
		prometheus.MustNewConstMetric(mMem, prometheus.GaugeValue, float64(r.Results.Mem)/100.0),
		prometheus.MustNewConstMetric(mDiskUsage, prometheus.GaugeValue, float64(r.Results.DiskUsage)/100.0),
		prometheus.MustNewConstMetric(mSessionCount, prometheus.GaugeValue, float64(r.Results.SessionCount)),
		prometheus.MustNewConstMetric(mConnCntPerSec, prometheus.GaugeValue, float64(r.Results.ConnCntPerSec)),
		prometheus.MustNewConstMetric(mLogDisk, prometheus.GaugeValue, 1, r.Results.LogDisk),
		prometheus.MustNewConstMetric(mDBStatus, prometheus.GaugeValue, 1, r.Results.DBStatus),
	}
	return m, true
}
