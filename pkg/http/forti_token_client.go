// Package http implements a small HTTP client for the FortiWeb REST API
// using API token authentication.
//
// # Copyright (C) 2025
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/mlueckert/fortiweb_exporter/internal/config"
)

// HTTPClient is the subset of *http.Client used by fortiTokenClient, useful
// for stubbing out in tests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// FortiHTTP is the interface used by probes to talk to a FortiWeb device.
type FortiHTTP interface {
	Get(path string, query string, obj interface{}) error
}

type fortiTokenClient struct {
	tgt url.URL
	hc  HTTPClient
	ctx context.Context
	tok config.Token
}

func (c *fortiTokenClient) newGetRequest(url string) (*http.Request, error) {
	r, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Authorization", string(c.tok))
	return r, nil
}

func (c *fortiTokenClient) Get(path string, query string, obj interface{}) error {
	u := c.tgt
	u.Path = path
	u.RawQuery = query

	req, err := c.newGetRequest(u.String())
	if err != nil {
		return err
	}

	req = req.WithContext(c.ctx)
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// FortiWeb answers some endpoints with non-standard 2xx codes (e.g. 220
	// for system/status.hastatus), so accept the whole 2xx range.
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("response code was %d, expected 2xx (path: %q, body: %q)", resp.StatusCode, path, truncate(b, 256))
	}

	if err := json.Unmarshal(b, obj); err != nil {
		return fmt.Errorf("failed to decode response (code %d, path: %q, body: %q): %v", resp.StatusCode, path, truncate(b, 256), err)
	}
	return nil
}

func truncate(b []byte, n int) string {
	if len(b) > n {
		return string(b[:n]) + "..."
	}
	return string(b)
}

func (c *fortiTokenClient) String() string {
	return c.tgt.String()
}

func newFortiTokenClient(ctx context.Context, tgt url.URL, hc HTTPClient, token config.Token) (*fortiTokenClient, error) {
	return &fortiTokenClient{tgt, hc, ctx, token}, nil
}

// EncodeToken builds a FortiWeb API token: the base64 encoding of
// {"username":"...","password":"...","vdom":"..."}. An empty vdom defaults
// to config.DefaultVdom.
func EncodeToken(username, password, vdom string) (config.Token, error) {
	if vdom == "" {
		vdom = config.DefaultVdom
	}
	b, err := json.Marshal(struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Vdom     string `json:"vdom"`
	}{username, password, vdom})
	if err != nil {
		return "", err
	}
	return config.Token(base64.StdEncoding.EncodeToString(b)), nil
}

// ResolveToken returns auth.Token if set, otherwise a token encoded from
// auth.Username, auth.Password and auth.Vdom.
func ResolveToken(auth config.TargetAuth) (config.Token, error) {
	if auth.Token != "" {
		return auth.Token, nil
	}
	if auth.Username == "" || auth.Password == "" {
		return "", fmt.Errorf("either token or username and password are required")
	}
	return EncodeToken(auth.Username, auth.Password, auth.Vdom)
}

// NewFortiClient creates a FortiHTTP client for tgt using the given
// authentication data.
func NewFortiClient(ctx context.Context, tgt url.URL, hc *http.Client, auth config.TargetAuth) (FortiHTTP, error) {
	if tgt.Scheme != "https" {
		return nil, fmt.Errorf("FortiWeb only supports token authentication for HTTPS connections")
	}

	token, err := ResolveToken(auth)
	if err != nil {
		return nil, fmt.Errorf("invalid authentication data for %q: %v", tgt.String(), err)
	}

	return newFortiTokenClient(ctx, tgt, hc, token)
}

// Configure applies TLS settings (extra trusted CAs, handshake timeout,
// insecure mode) to http.DefaultTransport.
func Configure(aConfig config.FortiWebExporterConfig) error {
	roots, err := x509.SystemCertPool()
	if err != nil {
		log.Fatalf("Unable to fetch system CA store: %v", err)
		return err
	}
	for _, cert := range aConfig.TlsExtraCAs {
		if ok := roots.AppendCertsFromPEM(cert.Content); !ok {
			return fmt.Errorf("failed to append certs from PEM %q, unknown error", cert.Path)
		}
	}
	tc := &tls.Config{RootCAs: roots}
	if aConfig.TLSInsecure {
		tc.InsecureSkipVerify = true
	}
	http.DefaultTransport.(*http.Transport).TLSHandshakeTimeout = time.Duration(aConfig.TLSTimeout) * time.Second
	http.DefaultTransport.(*http.Transport).TLSClientConfig = tc
	return nil
}
