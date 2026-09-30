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

	"errors"

	fortiHTTP "github.com/mlueckert/fortiweb_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSystemResource(t *testing.T) {
	c := newFakeClient()
	c.prepare("api/v2.0/system/status.systemresource", "testdata/system_status_systemresource.jsonnet")
	r := prometheus.NewPedanticRegistry()
	if !testProbe(probeSystemResource, c, r) {
		t.Errorf("probeSystemResource() returned non-success")
	}

	em := `
	# HELP fortiweb_connections_per_second Current amount of new connections per second, as reported by system/status.systemresource
	# TYPE fortiweb_connections_per_second gauge
	fortiweb_connections_per_second 0
	# HELP fortiweb_cpu_usage_ratio Current CPU usage ratio, as reported by system/status.systemresource
	# TYPE fortiweb_cpu_usage_ratio gauge
	fortiweb_cpu_usage_ratio 0.05
	# HELP fortiweb_current_sessions Current amount of sessions, as reported by system/status.systemresource
	# TYPE fortiweb_current_sessions gauge
	fortiweb_current_sessions 15
	# HELP fortiweb_db_status_info Status of the local database, as reported by system/status.systemresource
	# TYPE fortiweb_db_status_info gauge
	fortiweb_db_status_info{status="Available"} 1
	# HELP fortiweb_disk_usage_ratio Current disk usage ratio, as reported by system/status.systemresource
	# TYPE fortiweb_disk_usage_ratio gauge
	fortiweb_disk_usage_ratio 0.7
	# HELP fortiweb_log_disk_info Status of the log disk, as reported by system/status.systemresource
	# TYPE fortiweb_log_disk_info gauge
	fortiweb_log_disk_info{status="Available"} 1
	# HELP fortiweb_memory_usage_ratio Current memory usage ratio, as reported by system/status.systemresource
	# TYPE fortiweb_memory_usage_ratio gauge
	fortiweb_memory_usage_ratio 0.92
	`

	if err := testutil.GatherAndCompare(r, strings.NewReader(em)); err != nil {
		t.Fatalf("metric compare: err %v", err)
	}
}

func TestSystemResourceError(t *testing.T) {
	r := prometheus.NewPedanticRegistry()
	bc := &brokenClient{}
	if testProbe(probeSystemResource, bc, r) {
		t.Errorf("probeSystemResource() returned success, expected failure")
	}
}

type brokenClient struct{}

func (b *brokenClient) Get(path string, query string, obj interface{}) error {
	return errors.New("simulated error")
}

func (b *brokenClient) WithVdom(string) (fortiHTTP.FortiHTTP, error) {
	return b, nil
}
