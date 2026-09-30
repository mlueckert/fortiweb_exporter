// Copyright (C) 2025
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package probe

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mlueckert/fortiweb_exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ProbeHandler is the http.HandlerFunc for the /probe endpoint. It expects
// "target" (required), "profile", "username" and "password" query parameters.
func ProbeHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	paramMap := make(map[string]string)
	target := params.Get("target")
	paramMap["target"] = params.Get("target")
	if _, supplied := params["token"]; supplied {
		http.Error(w, "Token parameter is not supported; supply username and password", http.StatusBadRequest)
		return
	}
	if _, supplied := params["vdom"]; supplied {
		http.Error(w, "VDOM parameter is not supported; policy VDOMs are discovered automatically", http.StatusBadRequest)
		return
	}
	for _, k := range []string{"username", "password", "profile"} {
		if v := params.Get(k); v != "" {
			paramMap[k] = v
		}
	}

	if target == "" {
		http.Error(w, "Target parameter missing or empty", http.StatusBadRequest)
		return
	}
	savedConfig := config.GetConfig()
	probeSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_success",
		Help: "Whether or not the probe succeeded",
	})
	probeDurationGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_duration_seconds",
		Help: "How many seconds the probe took to complete",
	})
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(savedConfig.ScrapeTimeout)*time.Second)
	defer cancel()
	registry := prometheus.NewRegistry()
	registry.MustRegister(probeSuccessGauge)
	registry.MustRegister(probeDurationGauge)
	start := time.Now()
	pc := &ProbeCollector{}
	registry.MustRegister(pc)
	success, err := pc.Probe(ctx, paramMap, &http.Client{}, savedConfig)
	if err != nil {
		log.Printf("Probe request rejected; error is: %v", err)
		http.Error(w, fmt.Sprintf("probe: %v", err), http.StatusBadRequest)
		return
	}
	duration := time.Since(start).Seconds()
	probeDurationGauge.Set(duration)
	if success {
		probeSuccessGauge.Set(1)
	} else {
		// probeSuccessGauge default is 0
		log.Printf("Probe of %q failed, took %.3f seconds", target, duration)
	}
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
