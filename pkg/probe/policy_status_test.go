// Copyright (C) 2025
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package probe

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestPolicyStatus(t *testing.T) {
	c := newFakeClient()
	c.prepare("api/v2.0/policy/policystatus", "testdata/policy_policyStatus.jsonnet")
	r := prometheus.NewPedanticRegistry()
	if !testProbe(probePolicyStatus, c, r) {
		t.Fatalf("probePolicyStatus() returned non-success")
	}

	mfs, err := r.Gather()
	if err != nil {
		t.Fatal(err)
	}

	// Values of policy "pol-app01" in the testdata.
	want := map[string]float64{
		"fortiweb_policy_sessions":               0,
		"fortiweb_policy_connections_per_second": 0,
		"fortiweb_policy_client_rtt":             1,
		"fortiweb_policy_server_rtt":             6,
		"fortiweb_policy_app_response_time":      56,
	}
	wantLabels := map[string]string{
		"policy":     "17896",
		"name":       "pol-app01",
		"status":     "enable",
		"protocol":   "HTTP",
		"http_port":  "80",
		"https_port": "443",
		"mode":       "Single Server/Server Pool",
	}

	if len(mfs) != len(want) {
		t.Fatalf("got %d metric families, want %d", len(mfs), len(want))
	}
	for _, mf := range mfs {
		v, ok := want[mf.GetName()]
		if !ok {
			t.Errorf("unexpected metric %q", mf.GetName())
			continue
		}
		if n := len(mf.GetMetric()); n != 26 {
			t.Errorf("%s: got %d series, want 26", mf.GetName(), n)
		}
		found := false
		for _, m := range mf.GetMetric() {
			labels := map[string]string{}
			for _, lp := range m.GetLabel() {
				labels[lp.GetName()] = lp.GetValue()
			}
			if labels["name"] != wantLabels["name"] {
				continue
			}
			found = true
			for k, lv := range wantLabels {
				if labels[k] != lv {
					t.Errorf("%s: label %s = %q, want %q", mf.GetName(), k, labels[k], lv)
				}
			}
			if len(labels) != len(wantLabels) {
				t.Errorf("%s: got labels %v, want %v", mf.GetName(), labels, wantLabels)
			}
			if got := m.GetGauge().GetValue(); got != v {
				t.Errorf("%s: got %v, want %v", mf.GetName(), got, v)
			}
		}
		if !found {
			t.Errorf("%s: series for %q not found", mf.GetName(), wantLabels["name"])
		}
	}
}

func TestPolicyStatusError(t *testing.T) {
	r := prometheus.NewPedanticRegistry()
	if testProbe(probePolicyStatus, &brokenClient{}, r) {
		t.Errorf("probePolicyStatus() returned success, expected failure")
	}
}
