// Server executable of fortiweb_exporter
//
// # Copyright (C) 2025
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/mlueckert/fortiweb_exporter/internal/config"
	fortiHTTP "github.com/mlueckert/fortiweb_exporter/pkg/http"
	"github.com/mlueckert/fortiweb_exporter/pkg/probe"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Version = "(devel)"
	GitHash = "(no hash)"
)

func setUpMetricsEndpoint() {
	fortiwebExporterInfo := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fortiweb_exporter_build_info",
		Help: "This info metric contains build information for about the exporter",
	}, []string{"version", "revision"})

	fortiwebExporterInfo.With(prometheus.Labels{
		"version":  strings.TrimPrefix(Version, "v"),
		"revision": GitHash,
	}).Set(1)
}

// ipRestrictionMiddleware is a middleware function that restricts access to
// HTTP handlers based on the client's IP address. It checks if the client's IP
// address is within the allowed subnets. If the IP address is not allowed, it
// responds with a "Forbidden" status.
//
// Parameters:
// - next: The next http.Handler to be called if the IP address is allowed.
// - allowedSubnets: A slice of strings representing the allowed IP subnets.
//
// Returns:
// - An http.Handler that wraps the next handler with IP restriction logic.
func ipRestrictionMiddleware(next http.Handler, allowedSubnets []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]
		allowed := false
		for _, subnet := range allowedSubnets {
			if strings.HasPrefix(ip, subnet) || subnet == "0.0.0.0" {
				allowed = true
				break
			}
		}
		if !allowed {
			http.Error(w, "Forbidden, check allowed_subnets", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	log.Printf("FortiwebExporter %s ( %s )", Version, GitHash)
	setUpMetricsEndpoint()

	if err := config.Init(); err != nil {
		log.Fatalf("Initialization error: %+v", err)
	}

	savedConfig := config.GetConfig()

	if err := fortiHTTP.Configure(savedConfig); err != nil {
		log.Fatalf("%+v", err)
	}

	http.Handle("/metrics", ipRestrictionMiddleware(promhttp.Handler(), savedConfig.AllowedSubnets))
	http.Handle("/probe", ipRestrictionMiddleware(http.HandlerFunc(probe.ProbeHandler), savedConfig.AllowedSubnets))

	log.Printf("Fortiweb exporter running, listening on %q", savedConfig.Listen)
	if err := http.ListenAndServe(savedConfig.Listen, nil); err != nil {
		log.Fatalf("Unable to serve: %v", err)
	}
}
