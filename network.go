// Copyright 2017 Jeff Foley. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	asnServer = "asn.shadowserver.org"
	asnPort   = 43
)

type ASRecord struct {
	ASN    int
	Prefix string
	ASName string
	CN     string
	ISP    string
}

func IPToASRecord(ip string) (*ASRecord, error) {
	dialString := fmt.Sprintf("%s:%d", asnServer, asnPort)

	conn, err := net.Dial("tcp", dialString)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to: %s", dialString)
	}
	defer conn.Close()

	fmt.Fprintf(conn, "begin origin\n%s\nend\n", ip)
	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("Failed to read origin response for IP: %s", ip)
	}

	record := parseOriginResponse(line)
	if record == nil {
		return nil, fmt.Errorf("Failed to parse origin response for IP: %s", ip)
	}

	return record, nil
}

func ASNToNetblocks(asn int) ([]string, error) {
	dialString := fmt.Sprintf("%s:%d", asnServer, asnPort)

	conn, err := net.Dial("tcp", dialString)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to: %s", dialString)
	}
	defer conn.Close()

	fmt.Fprintf(conn, "prefix %d\n", asn)
	reader := bufio.NewReader(conn)

	var blocks []string

	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			blocks = append(blocks, strings.TrimSpace(line))
		}

		if err != nil {
			break
		}
	}

	if len(blocks) == 0 {
		return nil, fmt.Errorf("No netblocks returned for AS%d", asn)
	}
	return blocks, nil
}

func IPToCIDR(addr string) (*ASRecord, *net.IPNet, error) {
	// Get the AS record for the IP address
	record, err := IPToASRecord(addr)
	if err != nil {
		return nil, nil, err
	}
	// Get the netblocks associated with the ASN
	netblocks, err := ASNToNetblocks(record.ASN)
	if err != nil {
		return nil, nil, err
	}
	// Convert the CIDR into Go net types, and select the correct netblock
	var cidr *net.IPNet
	ip := net.ParseIP(addr)
	for _, nb := range netblocks {
		_, ipnet, err := net.ParseCIDR(nb)

		if err == nil && ipnet.Contains(ip) {
			cidr = ipnet
			break
		}
	}

	if cidr != nil {
		return record, cidr, nil
	}
	return nil, nil, errors.New("The IP address did not belong within the netblocks")
}

/* Private functions */

func parseOriginResponse(line string) *ASRecord {
	fields := strings.Split(line, " | ")

	asn, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil
	}

	return &ASRecord{
		ASN:    asn,
		Prefix: strings.TrimSpace(fields[2]),
		ASName: strings.TrimSpace(fields[3]),
		CN:     strings.TrimSpace(fields[4]),
		ISP:    strings.TrimSpace(fields[5]),
	}
}
