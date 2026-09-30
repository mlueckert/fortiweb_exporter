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
	"strings"

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

type systemVIP struct {
	Domains string `json:"domains"`
}

type systemVIPResponse struct {
	Results []systemVIP `json:"results"`
}

func probePolicyStatus(c http.FortiHTTP) ([]prometheus.Metric, bool) {
	labels := []string{"vdom", "policy", "name", "status", "protocol", "http_port", "https_port", "mode"}
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

	var vdomResponse systemVIPResponse
	if err := c.Get("api/v2.0/system/vip", "", &vdomResponse); err != nil {
		log.Printf("Error: %v", err)
		return nil, false
	}

	vdoms := make([]string, 0)
	seenVDOMs := make(map[string]struct{})
	for _, vip := range vdomResponse.Results {
		for _, vdom := range strings.Fields(vip.Domains) {
			if _, exists := seenVDOMs[vdom]; exists {
				continue
			}
			seenVDOMs[vdom] = struct{}{}
			vdoms = append(vdoms, vdom)
		}
	}
	if len(vdoms) == 0 {
		log.Printf("Error: system/vip returned no VDOMs in domains")
		return nil, false
	}

	success := true
	m := []prometheus.Metric{}
	for _, vdom := range vdoms {
		vdomClient, err := c.WithVdom(vdom)
		if err != nil {
			log.Printf("Error: unable to create client for VDOM %q: %v", vdom, err)
			success = false
			continue
		}

		var r policyStatusResponse
		if err := vdomClient.Get("api/v2.0/policy/policystatus", "", &r); err != nil {
			log.Printf("Error: unable to get policy status for VDOM %q: %v", vdom, err)
			success = false
			continue
		}

		for _, p := range r.Results {
			lv := []string{vdom, strconv.Itoa(p.Policy), p.Name, p.Status, p.Protocol, p.HTTPPort, p.HTTPSPort, p.Mode}
			m = append(m,
				prometheus.MustNewConstMetric(mSessionCount, prometheus.GaugeValue, float64(p.SessionCount), lv...),
				prometheus.MustNewConstMetric(mConnCntPerSec, prometheus.GaugeValue, float64(p.ConnCntPerSec), lv...),
				prometheus.MustNewConstMetric(mClientRTT, prometheus.GaugeValue, float64(p.ClientRTT), lv...),
				prometheus.MustNewConstMetric(mServerRTT, prometheus.GaugeValue, float64(p.ServerRTT), lv...),
				prometheus.MustNewConstMetric(mAppResponseTime, prometheus.GaugeValue, float64(p.AppResponseTime), lv...),
			)
		}
	}
	return m, success
}
