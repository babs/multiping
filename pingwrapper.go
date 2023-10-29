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

var re_host_w_proto = regexp.MustCompile(`^(tcp)://(\[?.+?\]?):(\d+)$`)

func NewPingWrapper(host string, options Options, transition_writer *TransitionWriter) PingWrapperInterface {

	host_findings := re_host_w_proto.FindAllStringSubmatch(host, -1)
	if len(host_findings) > 0 {
		port, err := strconv.Atoi(host_findings[0][3])
		if err != nil {
			log.Fatal(err)
		}
		return &TCPPingWrapper{
			host:  host_findings[0][2],
			ip:    mustResolve(host_findings[0][2]),
			port:  port,
			stats: &PWStats{transition_writer: transition_writer},
		}
	} else if *options.system {
		return &SystemPingWrapper{
			host:  host,
			ip:    mustResolve(host),
			stats: &PWStats{transition_writer: transition_writer},
		}
	} else {
		return &ProbingWrapper{
			host:       host,
			ip:         mustResolve(host),
			privileged: *options.privileged,
			stats:      &PWStats{transition_writer: transition_writer},
		}
	}
}

func mustResolve(host string) *net.IPAddr {
	host = strings.Trim(host, "[]")
	ipaddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		log.Fatal(err)
	}
	return ipaddr
}
