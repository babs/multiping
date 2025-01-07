package main

import (
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type PingWrapperInterface interface {
	Start()
	Stop()
	Host() string
	CalcStats(int64) PWStats
}

var re_host_w_proto = regexp.MustCompile(`^(tcp|ip)([46])?://(\[?.+?\]?)(?::(\d+))?$`)

func NewPingWrapper(host string, options Options, transition_writer *TransitionWriter) PingWrapperInterface {

	host_findings := re_host_w_proto.FindAllStringSubmatch(host, -1)

	var found_proto, found_ip_family, found_host, found_port string
	var found_port_int int

	if len(host_findings) > 0 {
		found_proto = host_findings[0][1]
		found_ip_family = host_findings[0][2]
		found_host = host_findings[0][3]
		found_port = host_findings[0][4]
	} else {
		found_host = host
	}

	if found_proto == "tcp" {

		if found_port == "" {
			log.Fatalf("%v: tcp probing requested but no port given\n", host)
		}
		port, err := strconv.Atoi(found_port)
		if err != nil {
			log.Fatalf("%v: %v\n", host, err)
		}
		if port <= 0 || port > 65535 {
			log.Fatalf("%v: tcp probing port invalid: %v\n", host, port)
		}
		found_port_int = port

		return &TCPPingWrapper{
			host:  found_host,
			ip:    mustResolve(found_host, found_ip_family),
			port:  found_port_int,
			stats: &PWStats{transition_writer: transition_writer},
		}
	} else if *options.system {
		return &SystemPingWrapper{
			host:         host,
			ip:           mustResolve(found_host, found_ip_family),
			stats:        &PWStats{transition_writer: transition_writer},
			ping_options: *options.system_ping_options,
		}
	} else {
		return &ProbingWrapper{
			host:       host,
			ip:         mustResolve(found_host, found_ip_family),
			privileged: *options.privileged,
			size:       *options.size,
			stats:      &PWStats{transition_writer: transition_writer},
		}
	}
}

func mustResolve(host string, ip_family string) *net.IPAddr {
	host = strings.Trim(host, "[]")
	ipaddr, err := net.ResolveIPAddr("ip"+ip_family, host)
	if err != nil {
		log.Fatal(err)
	}
	return ipaddr
}
