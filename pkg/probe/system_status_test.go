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

func TestSystemStatus(t *testing.T) {
	c := newFakeClient()
	c.prepare("api/v2.0/system/status.systemstatus", "testdata/system_status_systemstatus.jsonnet")
	r := prometheus.NewPedanticRegistry()
	if !testProbe(probeSystemStatus, c, r) {
		t.Fatalf("probeSystemStatus() returned non-success")
	}

	// 171d 4h 49m = 14791740s
	em := `
	# HELP fortiweb_system_info System information, as reported by system/status.systemstatus
	# TYPE fortiweb_system_info gauge
	fortiweb_system_info{cluster_name="fortiweb01",cluster_role="Primary",firmware_version="FortiWeb-Azure 7.6.7,build1111(GA.M),260204",serial="FWBVM00000000001"} 1
	# HELP fortiweb_uptime_seconds Time since the system was started in seconds (minute resolution), as reported by system/status.systemstatus
	# TYPE fortiweb_uptime_seconds gauge
	fortiweb_uptime_seconds 14791740
	`

	if err := testutil.GatherAndCompare(r, strings.NewReader(em)); err != nil {
		t.Fatalf("metric compare: err %v", err)
	}
}

func TestSystemStatusError(t *testing.T) {
	r := prometheus.NewPedanticRegistry()
	if testProbe(probeSystemStatus, &brokenClient{}, r) {
		t.Errorf("probeSystemStatus() returned success, expected failure")
	}
}
