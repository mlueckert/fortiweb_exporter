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

// policyStatusResult is a single entry of the FortiWeb monitor API endpoint
// api/v2.0/policy/policystatus.
type policyStatusResult struct {
	Policy          int    `json:"policy"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	Protocol        string `json:"protocol"`
	HTTPPort        string `json:"httpPort"`
	HTTPSPort       string `json:"httpsPort"`
	Mode            string `json:"mode"`
	SessionCount    int    `json:"sessionCount"`
	ConnCntPerSec   int    `json:"connCntPerSec"`
	ClientRTT       int    `json:"client_rtt"`
	ServerRTT       int    `json:"server_rtt"`
	AppResponseTime int    `json:"app_response_time"`
}

type policyStatusResponse struct {
	Results []policyStatusResult `json:"results"`
}

func probePolicyStatus(c http.FortiHTTP) ([]prometheus.Metric, bool) {
	labels := []string{"policy", "name", "status", "protocol", "http_port", "https_port", "mode"}
	var (
		mSessionCount = prometheus.NewDesc(
			"fortiweb_policy_sessions",
			"Current amount of sessions of the policy, as reported by policy/policystatus",
			labels, nil,
		)
		mConnCntPerSec = prometheus.NewDesc(
			"fortiweb_policy_connections_per_second",
			"Current amount of new connections per second of the policy, as reported by policy/policystatus",
			labels, nil,
		)
		mClientRTT = prometheus.NewDesc(
			"fortiweb_policy_client_rtt",
			"Client round trip time of the policy, as reported by policy/policystatus",
			labels, nil,
		)
		mServerRTT = prometheus.NewDesc(
			"fortiweb_policy_server_rtt",
			"Server round trip time of the policy, as reported by policy/policystatus",
			labels, nil,
		)
		mAppResponseTime = prometheus.NewDesc(
			"fortiweb_policy_app_response_time",
			"Application response time of the policy, as reported by policy/policystatus",
			labels, nil,
		)
	)

	var r policyStatusResponse
	if err := c.Get("api/v2.0/policy/policystatus", "", &r); err != nil {
		log.Printf("Error: %v", err)
		return nil, false
	}

	m := []prometheus.Metric{}
	for _, p := range r.Results {
		lv := []string{strconv.Itoa(p.Policy), p.Name, p.Status, p.Protocol, p.HTTPPort, p.HTTPSPort, p.Mode}
		m = append(m,
			prometheus.MustNewConstMetric(mSessionCount, prometheus.GaugeValue, float64(p.SessionCount), lv...),
			prometheus.MustNewConstMetric(mConnCntPerSec, prometheus.GaugeValue, float64(p.ConnCntPerSec), lv...),
			prometheus.MustNewConstMetric(mClientRTT, prometheus.GaugeValue, float64(p.ClientRTT), lv...),
			prometheus.MustNewConstMetric(mServerRTT, prometheus.GaugeValue, float64(p.ServerRTT), lv...),
			prometheus.MustNewConstMetric(mAppResponseTime, prometheus.GaugeValue, float64(p.AppResponseTime), lv...),
		)
	}
	return m, true
}
