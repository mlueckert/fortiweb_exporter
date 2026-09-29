// Copyright (C) 2025
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package probe

import (
	"log"
	"strconv"

	"github.com/mlueckert/fortiweb_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

// systemOperationInterface is a network interface entry of the FortiWeb
// monitor API endpoint api/v2.0/system/status.systemoperation.
type systemOperationInterface struct {
	Name        string `json:"name"`
	Alias       string `json:"alias"`
	IPNetmask   string `json:"ip_netmask"`
	SpeedDuplex string `json:"speedDuplex"`
	TX          string `json:"tx"`
	RX          string `json:"rx"`
	TXBytes     string `json:"tx_bytes"`
	RXBytes     string `json:"rx_bytes"`
	Link        string `json:"link"`
}

type systemOperationResponse struct {
	Results struct {
		Network []systemOperationInterface `json:"network"`
	} `json:"results"`
}

func probeSystemInterface(c http.FortiHTTP) ([]prometheus.Metric, bool) {
	labels := []string{"interface"}
	var (
		mInfo = prometheus.NewDesc(
			"fortiweb_interface_info",
			"Network interface information, as reported by system/status.systemoperation",
			[]string{"interface", "alias", "ip_netmask", "speed_duplex"}, nil,
		)
		mLinkUp = prometheus.NewDesc(
			"fortiweb_interface_link_up",
			"Whether the link of the network interface is up (1) or not (0), as reported by system/status.systemoperation",
			labels, nil,
		)
		mTXPackets = prometheus.NewDesc(
			"fortiweb_interface_transmit_packets_total",
			"Number of packets transmitted on the network interface, as reported by system/status.systemoperation",
			labels, nil,
		)
		mRXPackets = prometheus.NewDesc(
			"fortiweb_interface_receive_packets_total",
			"Number of packets received on the network interface, as reported by system/status.systemoperation",
			labels, nil,
		)
		mTXBytes = prometheus.NewDesc(
			"fortiweb_interface_transmit_bytes_total",
			"Number of bytes transmitted on the network interface, as reported by system/status.systemoperation",
			labels, nil,
		)
		mRXBytes = prometheus.NewDesc(
			"fortiweb_interface_receive_bytes_total",
			"Number of bytes received on the network interface, as reported by system/status.systemoperation",
			labels, nil,
		)
	)

	var r systemOperationResponse
	if err := c.Get("api/v2.0/system/status.systemoperation", "", &r); err != nil {
		log.Printf("Error: %v", err)
		return nil, false
	}

	success := true
	m := []prometheus.Metric{}
	for _, i := range r.Results.Network {
		linkUp := 0.0
		if i.Link == "Up" {
			linkUp = 1
		}
		m = append(m,
			prometheus.MustNewConstMetric(mInfo, prometheus.GaugeValue, 1, i.Name, i.Alias, i.IPNetmask, i.SpeedDuplex),
			prometheus.MustNewConstMetric(mLinkUp, prometheus.GaugeValue, linkUp, i.Name),
		)
		for _, ctr := range []struct {
			desc *prometheus.Desc
			v    string
		}{
			{mTXPackets, i.TX},
			{mRXPackets, i.RX},
			{mTXBytes, i.TXBytes},
			{mRXBytes, i.RXBytes},
		} {
			v, err := strconv.ParseUint(ctr.v, 10, 64)
			if err != nil {
				log.Printf("Error: unable to parse counter of interface %q from system/status.systemoperation: %v", i.Name, err)
				success = false
				continue
			}
			m = append(m, prometheus.MustNewConstMetric(ctr.desc, prometheus.CounterValue, float64(v), i.Name))
		}
	}
	return m, success
}
