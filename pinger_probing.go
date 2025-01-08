package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type ProbingWrapper struct {
	host       string
	ip         *net.IPAddr
	hstring    string
	pinger     *probing.Pinger
	size       int
	stats      *PWStats
	privileged bool
}

func (w *ProbingWrapper) Start() {
	var err error
	w.pinger, err = probing.NewPinger(w.ip.String())
	if err != nil {
		log.Fatalf("pinger initialization failed %s, %s", w.host, err)
	}

	w.pinger.RecordRtts = false
	w.pinger.OnSend = w.onSend
	// pinger.OnSend = pingwrapper.OnRecv
	w.pinger.OnRecv = w.onRecv
	w.pinger.OnDuplicateRecv = w.onDuplicateRecv
	w.pinger.Size = w.size
	if runtime.GOOS == "linux" {
		w.pinger.SetDoNotFragment(true)
	}

	if runtime.GOOS == "windows" || os.Getuid() == 0 {
		w.pinger.SetPrivileged(true)
	} else {
		w.pinger.SetPrivileged(w.privileged)
	}

	w.hstring = fmt.Sprintf("%s (%s)", w.host, w.ip.String())

	w.stats.hrepr = w.host
	w.stats.iprepr = w.ip.IP.String()

	go func(w *ProbingWrapper) {
		err := w.pinger.Run()
		if err != nil {
			log.Fatalf("%s", err)
		}
	}(w)
}

func (w *ProbingWrapper) Stop() {
	w.pinger.Stop()
}

func (w *ProbingWrapper) onSend(pkt *probing.Packet) {
	w.stats.lastsent = time.Now().UnixNano()
}

func (w *ProbingWrapper) onRecv(pkt *probing.Packet) {
	// p.lastread = fmt.Sprintf("%d bytes from %s (%s): icmp_seq=%d time=%v",
	//	pkt.Nbytes, p.host, pkt.IPAddr, pkt.Seq, pkt.Rtt)
	// fmt.Print(p.lastread)
	w.stats.has_ever_received = true
	w.stats.lastrecv = time.Now().UnixNano()
	w.stats.lastrtt = pkt.Rtt
	w.stats.lastrtt_as_string = round(w.stats.lastrtt, 2).String()
}

func (w *ProbingWrapper) onDuplicateRecv(pkt *probing.Packet) {
	// p.lastread = fmt.Sprintf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v (DUP!)", pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.TTL)
}

func (w *ProbingWrapper) Host() string {
	return w.hstring
}

func (w *ProbingWrapper) CalcStats(timeout_threshold int64) PWStats {
	w.stats.ComputeState(timeout_threshold)
	return *w.stats
}

var divs = []time.Duration{
	time.Duration(1), time.Duration(10), time.Duration(100), time.Duration(1000)}

func round(d time.Duration, digits int) time.Duration {
	switch {
	case d > time.Second:
		d = d.Round(time.Second / divs[digits])
	case d > time.Millisecond:
		d = d.Round(time.Millisecond / divs[digits])
	case d > time.Microsecond:
		d = d.Round(time.Microsecond / divs[digits])
	}
	return d
}
