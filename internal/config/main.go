// Package config handles command-line flags and the API authentication map
// file used to configure fortiweb_exporter.
//
// # Copyright (C) 2025
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package config

import (
	"flag"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// FortiWebExporterParameter holds the raw command-line flags before parsing.
type FortiWebExporterParameter struct {
	AuthFile       *string
	Listen         *string
	ScrapeTimeout  *int
	TLSTimeout     *int
	TLSInsecure    *bool
	TlsExtraCAs    *string
	AllowedSubnets *string
}

// FortiWebExporterConfig holds the fully parsed exporter configuration.
type FortiWebExporterConfig struct {
	AuthKeys       AuthKeys
	Listen         string
	ScrapeTimeout  int
	TLSTimeout     int
	TLSInsecure    bool
	TlsExtraCAs    []LocalCert
	AllowedSubnets []string
}

// AuthKeys maps a Target (a scrape target URL) or a profile name to its
// authentication and probe selection.
type AuthKeys map[Target]TargetAuth

// Target identifies either a scrape target URL (e.g.
// "https://fortiweb.example.com") or the name of a probe profile.
type Target string

// Token is a FortiWeb API access token, i.e. the base64 encoded JSON
// {"username":"...","password":"...","vdom":"..."} sent verbatim in the
// Authorization header.
type Token string

// DefaultVdom is used when no vdom is configured.
const DefaultVdom = "root"

// ProbeList is a list of probe name prefixes, e.g. "System/Resource".
type ProbeList []string

// Probes selects which probes should run for a given target/profile.
type Probes struct {
	Include ProbeList
	Exclude ProbeList
}

// TargetAuth is the authentication and probe selection for a Target.
//
// Either Token (an already encoded token) or Username/Password/Vdom can be
// given. If Token is empty, it is built from Username/Password/Vdom.
type TargetAuth struct {
	Token    Token
	Username string
	Password string
	Vdom     string
	Probes   Probes
}

// LocalCert is an extra CA certificate to trust for TLS connections.
type LocalCert struct {
	Path    string
	Content []byte
}

var (
	parameter = FortiWebExporterParameter{
		AuthFile:       flag.String("auth-file", "fortiweb-key.yaml", "file containing the authentication map to use when connecting to a FortiWeb device"),
		Listen:         flag.String("listen", ":9723", "address to listen on"),
		ScrapeTimeout:  flag.Int("scrape-timeout", 30, "max seconds to allow a scrape to take"),
		TLSTimeout:     flag.Int("https-timeout", 10, "TLS Handshake timeout in seconds"),
		TLSInsecure:    flag.Bool("insecure", false, "Allow insecure certificates"),
		TlsExtraCAs:    flag.String("extra-ca-certs", "", "comma-separated files containing extra PEMs to trust for TLS connections in addition to the system trust store"),
		AllowedSubnets: flag.String("allowed-subnets", "0.0.0.0", "comma-separated list of allowed ips or subnets. If 0.0.0.0, all IPs are allowed"),
	}

	savedConfig *FortiWebExporterConfig
)

// Init parses flags and the auth file once. Subsequent calls are no-ops.
func Init() error {
	// check if already parsed
	if savedConfig != nil {
		return nil
	}
	return ReInit()
}

// MustReInit is like ReInit but fatally logs on error.
func MustReInit() {
	if err := ReInit(); err != nil {
		log.Fatalf("config.ReInit failed: %+v", err)
	}
}

// ReInit (re-)parses flags and the auth file, overwriting any prior
// configuration.
func ReInit() error {
	flag.Parse()

	savedConfig = &FortiWebExporterConfig{
		Listen:        *parameter.Listen,
		ScrapeTimeout: *parameter.ScrapeTimeout,
		TLSTimeout:    *parameter.TLSTimeout,
		TLSInsecure:   *parameter.TLSInsecure,
	}

	// parse AuthKeys
	af, err := os.ReadFile(*parameter.AuthFile)
	if err != nil {
		log.Fatalf("Failed to read API authentication map file: %v", err)
		return err
	}

	if err := yaml.Unmarshal(af, &savedConfig.AuthKeys); err != nil {
		log.Fatalf("Failed to parse API authentication map file: %v", err)
		return err
	}

	log.Printf("Loaded %d API keys", len(savedConfig.AuthKeys))

	// parse AllowedSubnets
	for _, subnet := range strings.Split(*parameter.AllowedSubnets, ",") {
		if subnet == "" {
			continue
		}
		savedConfig.AllowedSubnets = append(savedConfig.AllowedSubnets, subnet)
	}

	// parse ExtraCAs
	for _, eca := range strings.Split(*parameter.TlsExtraCAs, ",") {
		if eca == "" {
			continue
		}

		certs, err := os.ReadFile(eca)
		if err != nil {
			log.Fatalf("Failed to read extra CA file %q: %v", eca, err)
			return err
		}

		certObject := LocalCert{
			Path:    eca,
			Content: certs,
		}
		savedConfig.TlsExtraCAs = append(savedConfig.TlsExtraCAs, certObject)
	}

	return nil
}

// GetConfig returns a copy of the currently loaded configuration.
func GetConfig() FortiWebExporterConfig {
	return *savedConfig
}
