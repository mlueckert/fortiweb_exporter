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

// haStatusResult is the result of the FortiWeb monitor API endpoint
// api/v2.0/system/status.hastatus.
type haStatusResult struct {
	CfgSyncState string `json:"cfg_sync_state"`
}

type haStatusResponse struct {
	Results haStatusResult `json:"results"`
}

func probeSystemHAStatus(c http.FortiHTTP) ([]prometheus.Metric, bool) {
	mCfgSyncState := prometheus.NewDesc(
		"fortiweb_ha_cfg_sync_state_info",
		"HA configuration synchronization state, as reported by system/status.hastatus",
		[]string{"state"}, nil,
	)

	var r haStatusResponse
	if err := c.Get("api/v2.0/system/status.hastatus", "", &r); err != nil {
		log.Printf("Error: %v", err)
		return nil, false
	}
	if r.Results.CfgSyncState == "" {
		log.Printf("Error: system/status.hastatus returned no cfg_sync_state")
		return nil, false
	}

	return []prometheus.Metric{
		prometheus.MustNewConstMetric(mCfgSyncState, prometheus.GaugeValue, 1, r.Results.CfgSyncState),
	}, true
}
