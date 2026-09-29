// Tests of fortiweb_exporter probes
//
// # Copyright (C) 2025
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package probe

import (
	"encoding/json"
	"log"

	"github.com/google/go-jsonnet"
	"github.com/mlueckert/fortiweb_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

// fakeClient is a minimal in-memory FortiHTTP implementation for tests.
type fakeClient struct {
	data map[string][]byte
}

func newFakeClient() *fakeClient {
	return &fakeClient{data: map[string][]byte{}}
}

// prepare registers the evaluated jsonnet file jfile as response for path.
func (c *fakeClient) prepare(path string, jfile string) {
	output, err := jsonnet.MakeVM().EvaluateFile(jfile)
	if err != nil {
		log.Fatalf("Failed to evaluate jsonnet %q: %v", jfile, err)
	}
	c.data[path] = []byte(output)
}

func (c *fakeClient) Get(path string, query string, obj interface{}) error {
	d, ok := c.data[path]
	if !ok {
		log.Fatalf("Tried to get unprepared URL %q", path)
	}
	return json.Unmarshal(d, obj)
}

type Registry interface {
	MustRegister(...prometheus.Collector)
}

type testProbeCollector struct {
	metrics []prometheus.Metric
}

func (p *testProbeCollector) Collect(c chan<- prometheus.Metric) {
	for _, m := range p.metrics {
		c <- m
	}
}

func (p *testProbeCollector) Describe(c chan<- *prometheus.Desc) {
}

func testProbe(pf probeFunc, c http.FortiHTTP, r Registry) bool {
	m, ok := pf(c)
	if !ok {
		return false
	}
	p := &testProbeCollector{metrics: m}
	r.MustRegister(p)
	return true
}
