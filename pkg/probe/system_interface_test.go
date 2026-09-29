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

func TestSystemInterface(t *testing.T) {
	c := newFakeClient()
	c.prepare("api/v2.0/system/status.systemoperation", "testdata/system_status_systemoperation.jsonnet")
	r := prometheus.NewPedanticRegistry()
	if !testProbe(probeSystemInterface, c, r) {
		t.Fatalf("probeSystemInterface() returned non-success")
	}

	em := `
	# HELP fortiweb_interface_info Network interface information, as reported by system/status.systemoperation
	# TYPE fortiweb_interface_info gauge
	fortiweb_interface_info{alias="",interface="port1",ip_netmask="192.0.2.4/26",speed_duplex="50000 Mbps/Full Duplex"} 1
	fortiweb_interface_info{alias="",interface="port2",ip_netmask="192.0.2.196/26",speed_duplex="50000 Mbps/Full Duplex"} 1
	# HELP fortiweb_interface_link_up Whether the link of the network interface is up (1) or not (0), as reported by system/status.systemoperation
	# TYPE fortiweb_interface_link_up gauge
	fortiweb_interface_link_up{interface="port1"} 1
	fortiweb_interface_link_up{interface="port2"} 1
	# HELP fortiweb_interface_receive_bytes_total Number of bytes received on the network interface, as reported by system/status.systemoperation
	# TYPE fortiweb_interface_receive_bytes_total counter
	fortiweb_interface_receive_bytes_total{interface="port1"} 795104961637
	fortiweb_interface_receive_bytes_total{interface="port2"} 25891792048
	# HELP fortiweb_interface_receive_packets_total Number of packets received on the network interface, as reported by system/status.systemoperation
	# TYPE fortiweb_interface_receive_packets_total counter
	fortiweb_interface_receive_packets_total{interface="port1"} 1905623549
	fortiweb_interface_receive_packets_total{interface="port2"} 116643074
	# HELP fortiweb_interface_transmit_bytes_total Number of bytes transmitted on the network interface, as reported by system/status.systemoperation
	# TYPE fortiweb_interface_transmit_bytes_total counter
	fortiweb_interface_transmit_bytes_total{interface="port1"} 751354001817
	fortiweb_interface_transmit_bytes_total{interface="port2"} 69582088571
	# HELP fortiweb_interface_transmit_packets_total Number of packets transmitted on the network interface, as reported by system/status.systemoperation
	# TYPE fortiweb_interface_transmit_packets_total counter
	fortiweb_interface_transmit_packets_total{interface="port1"} 1914178188
	fortiweb_interface_transmit_packets_total{interface="port2"} 133420325
	`

	if err := testutil.GatherAndCompare(r, strings.NewReader(em)); err != nil {
		t.Fatalf("metric compare: err %v", err)
	}
}

func TestSystemInterfaceError(t *testing.T) {
	r := prometheus.NewPedanticRegistry()
	if testProbe(probeSystemInterface, &brokenClient{}, r) {
		t.Errorf("probeSystemInterface() returned success, expected failure")
	}
}
