// Copyright (C) 2025
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package probe

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSystemHAStatus(t *testing.T) {
	c := newFakeClient()
	c.prepare("api/v2.0/system/status.hastatus", "testdata/system_status_hastatus.jsonnet")
	r := prometheus.NewPedanticRegistry()
	if !testProbe(probeSystemHAStatus, c, r) {
		t.Fatalf("probeSystemHAStatus() returned non-success")
	}

	em := `
	# HELP fortiweb_ha_cfg_sync_state_info HA configuration synchronization state, as reported by system/status.hastatus
	# TYPE fortiweb_ha_cfg_sync_state_info gauge
	fortiweb_ha_cfg_sync_state_info{state="In sync"} 1
	`

	if err := testutil.GatherAndCompare(r, strings.NewReader(em)); err != nil {
		t.Fatalf("metric compare: err %v", err)
	}
}

func TestSystemHAStatusError(t *testing.T) {
	r := prometheus.NewPedanticRegistry()
	if testProbe(probeSystemHAStatus, &brokenClient{}, r) {
		t.Errorf("probeSystemHAStatus() returned success, expected failure")
	}
}
