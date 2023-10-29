//go:build !windows

package main

// inspired from https://github.com/cloverstd/tcping/blob/master/ping/tcp/tcp.go

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	tcpshaker "github.com/tevino/tcp-shaker"
)

type TCPPingWrapper struct {
	host          string
	ip            *net.IPAddr
	hstring       string
	port          int
	str_tgt       string
	stats         *PWStats
	stopCheckLoop bool
	loopTicker    *time.Ticker
}

func (w *TCPPingWrapper) Start() {
	w.hstring = fmt.Sprintf("tcp://%v:%v (%v:%v)", w.host, w.port, w.ip.String(), w.port)
	w.stats.hrepr = fmt.Sprintf("tcp://%v:%v", w.host, w.port)
	w.stats.iprepr = w.ip.IP.String()

	if strings.Contains(w.stats.iprepr, ":") {
		w.str_tgt = fmt.Sprintf("[%v]:%v", w.ip.String(), w.port)
		w.hstring = fmt.Sprintf("tcp://%v:%v ([%v]:%v)", w.host, w.port, w.ip.String(), w.port)
	} else {
		w.str_tgt = fmt.Sprintf("%v:%v", w.ip.String(), w.port)
	}

	w.stopCheckLoop = false
	w.loopTicker = time.NewTicker(time.Second)

	go func(w *TCPPingWrapper) {
		for !w.stopCheckLoop {
			go func(t *TCPPingWrapper) {
				t.spawnChecker()
			}(w)
			<-w.loopTicker.C
		}
	}(w)

}

func (w *TCPPingWrapper) spawnChecker() {
	checker := tcpshaker.NewChecker()

	ctx, stopChecker := context.WithCancel(context.Background())
	defer stopChecker()
	go func() {
		if err := checker.CheckingLoop(ctx); err != nil {
			fmt.Println("checking loop stopped due to fatal error: ", err)
		}
	}()
	<-checker.WaitReady()
	start := time.Now()
	w.stats.lastsent = time.Now().UnixNano()
	err := checker.CheckAddr(w.str_tgt, time.Second)
	if err == nil {
		w.stats.has_ever_received = true
		w.stats.lastrecv = time.Now().UnixNano()
		w.stats.lastrtt = time.Since(start)
		w.stats.lastrtt_as_string = round(w.stats.lastrtt, 2).String()
	}
}

func (w *TCPPingWrapper) Stop() {
	w.stopCheckLoop = true
	w.loopTicker.Stop()
}

func (w *TCPPingWrapper) Host() string {
	return w.hstring
}

func (w *TCPPingWrapper) CalcStats(timeout_threshold int64) PWStats {
	w.stats.ComputeState(timeout_threshold)
	return *w.stats
}
