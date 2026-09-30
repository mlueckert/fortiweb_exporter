// All currently supported probes
//
// # Copyright (C) 2025
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package probe

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mlueckert/fortiweb_exporter/internal/config"
	fortiHTTP "github.com/mlueckert/fortiweb_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

// ProbeCollector collects the metrics gathered by a single Probe() call.
type ProbeCollector struct {
	metrics []prometheus.Metric
}

type probeFunc func(fortiHTTP.FortiHTTP) ([]prometheus.Metric, bool)

type probeDetailedFunc struct {
	name     string
	function probeFunc
}

// Probe connects to the target FortiWeb device given in target["target"],
// authenticating with target["username"] and target["password"] (falling back to the values of the profile named in
// target["profile"]), and runs all selected probes. System probes use the root
// VDOM; policy status is collected once for each VDOM discovered from system/vip.
func (p *ProbeCollector) Probe(ctx context.Context, target map[string]string, hc *http.Client, savedConfig config.FortiWebExporterConfig) (bool, error) {
	tgt, err := url.Parse(target["target"])
	if err != nil {
		return false, fmt.Errorf("url.Parse failed: %v", err)
	}

	if tgt.Scheme != "https" && tgt.Scheme != "http" {
		return false, fmt.Errorf("unsupported scheme %q", tgt.Scheme)
	}

	// Filter anything else than scheme and hostname
	u := url.URL{
		Scheme: tgt.Scheme,
		Host:   tgt.Host,
	}

	auth := resolveAuth(target, savedConfig.AuthKeys[config.Target(target["profile"])])

	c, err := fortiHTTP.NewFortiClient(ctx, u, hc, auth)
	if err != nil {
		return false, err
	}
	includedProbes := auth.Probes.Include
	excludedProbes := auth.Probes.Exclude

	success := true
	for _, aProbe := range []probeDetailedFunc{
		{"System/Resource", probeSystemResource},
		{"System/Status", probeSystemStatus},
		{"System/HAStatus", probeSystemHAStatus},
		{"System/Interface", probeSystemInterface},
		{"Policy/Status", probePolicyStatus},
	} {
		wanted := false

		if len(includedProbes) == 0 {
			wanted = true
		} else {
			for _, wantedProbe := range includedProbes {
				if strings.HasPrefix(aProbe.name, wantedProbe) {
					wanted = true
					break
				}
			}
		}

		if len(excludedProbes) != 0 {
			for _, unwantedProbe := range excludedProbes {
				if strings.HasPrefix(aProbe.name, unwantedProbe) {
					wanted = false
					break
				}
			}
		}

		if !wanted {
			continue
		}

		m, ok := aProbe.function(c)
		if !ok {
			success = false
		}
		p.metrics = append(p.metrics, m...)
	}

	return success, nil
}

// resolveAuth merges request credentials over the profile.
func resolveAuth(params map[string]string, profile config.TargetAuth) config.TargetAuth {
	auth := profile
	if v := params["username"]; v != "" {
		auth.Username = v
	}
	if v := params["password"]; v != "" {
		auth.Password = v
	}
	return auth
}

// Collect implements prometheus.Collector.
func (p *ProbeCollector) Collect(c chan<- prometheus.Metric) {
	for _, m := range p.metrics {
		c <- m
	}
}

// Describe implements prometheus.Collector.
func (p *ProbeCollector) Describe(c chan<- *prometheus.Desc) {
}
