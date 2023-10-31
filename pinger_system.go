package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type SystemPingWrapper struct {
	host    string
	ip      *net.IPAddr
	hstring string
	stats   *PWStats
	cmd     *exec.Cmd
}

var time_extractor = regexp.MustCompile(`time[=<]([\d\.]+) *(.?s)`)

func (w *SystemPingWrapper) Start() {
	w.hstring = fmt.Sprintf("%s (%s)", w.host, w.ip.String())
	w.stats.hrepr = w.host
	w.stats.iprepr = w.ip.String()

	var path string

	// Looks like an ipv6 ? search for ping6
	// Some systems doesn't have ping6 because ping handle both v4 and v6
	// so not finding ping6 is not necessarily a problem
	if strings.Contains(w.ip.String(), ":") {
		path, _ = exec.LookPath("ping6")
	}

	if path == "" {
		var err error
		path, err = exec.LookPath("ping")
		if err != nil {
			log.Fatal(err)
		}
	}

	args := make([]string, 0)

	if runtime.GOOS == "windows" {
		args = append(args, "-t")
	}
	args = append(args, w.ip.String())

	w.cmd = exec.Command(path, args...)

	w.stats = &PWStats{
		state: true,
	}
	r, _ := w.cmd.StdoutPipe()
	scanner := bufio.NewScanner(r)
	go func() {
		// Read line by line and process it
		for scanner.Scan() {
			line := scanner.Text()
			extracted := time_extractor.FindAllStringSubmatch(line, -1)
			if len(extracted) > 0 {
				w.stats.lastrecv = time.Now().UnixNano()
				w.stats.lastrtt_as_string = extracted[0][1] + extracted[0][2]
			}
		}
		w.stats.error_message = fmt.Sprintf("%v exited code %v", w.cmd.String(), w.cmd.ProcessState.ExitCode())
	}()
	w.cmd.Start()
}

func (w *SystemPingWrapper) Stop() {
	w.cmd.Process.Signal(os.Interrupt)
}

func (w *SystemPingWrapper) Host() string {
	return w.hstring
}

func (w *SystemPingWrapper) CalcStats(timeout_threshold int64) PWStats {
	w.stats.ComputeState(timeout_threshold)
	return *w.stats
}
