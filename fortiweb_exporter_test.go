// Tests of fortiweb_exporter
//
// # Copyright (C) 2025
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIPRestrictionMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		allowedSubnets []string
		remoteAddr     string
		expectedStatus int
	}{
		{
			name:           "Allowed IP",
			allowedSubnets: []string{"192.168.1."},
			remoteAddr:     "192.168.1.100:12345",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Forbidden IP",
			allowedSubnets: []string{"192.168.1."},
			remoteAddr:     "10.0.0.1:12345",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Allowed IP with 0.0.0.0",
			allowedSubnets: []string{"0.0.0.0"},
			remoteAddr:     "10.0.0.1:12345",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty allowed subnets",
			allowedSubnets: []string{},
			remoteAddr:     "10.0.0.1:12345",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := ipRestrictionMiddleware(nextHandler, tt.allowedSubnets)

			req := httptest.NewRequest("GET", "http://example.com", nil)
			req.RemoteAddr = tt.remoteAddr
			rr := httptest.NewRecorder()

			middleware.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusForbidden {
				expectedBody := "Forbidden, check allowed_subnets\n"
				if rr.Body.String() != expectedBody {
					t.Errorf("handler returned unexpected body: got %v want %v",
						rr.Body.String(), expectedBody)
				}
			}
		})
	}
}
