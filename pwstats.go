package main

import (
	"encoding/json"
	"strings"
	"time"
)

type PWStats struct {
	lastsent           int64
	lastrecv           int64
	lastrtt            time.Duration
	lastrtt_as_string  string
	last_loss_nano     int64
	last_loss_duration int64
	last_seen_nano     int64
	state              bool
	first_called       bool
	has_ever_received  bool
	startup_time       int64
	transition_writer  *TransitionWriter
	error_message      string
	hrepr              string
	iprepr             string
}

func (p *PWStats) ComputeState(timeout_threshold int64) {
	if p.startup_time == 0 {
		p.startup_time = time.Now().UnixNano()
	}
	old_last_seen := p.last_seen_nano
	p.last_seen_nano = time.Now().UnixNano() - p.lastrecv
	new_state := p.last_seen_nano < timeout_threshold
	// TODO: Algo to review completely

	if !p.state && new_state {
		if p.first_called {
			p.last_loss_nano = time.Now().UnixNano()
			p.last_loss_duration = old_last_seen
		} else {
			p.first_called = true
		}
	}
	if p.state != new_state {
		var sb strings.Builder
		now := time.Now()

		var transition string
		if new_state {
			transition = "down to up"
		} else {
			transition = "up to down"
		}

		jsonString, _ := json.Marshal(
			struct {
				Timestamp  string
				UnixNano   int64
				Host       string
				Ip         string
				Transition string
				State      bool
			}{
				now.String(),
				now.UnixNano(),
				p.hrepr,
				p.iprepr,
				transition,
				new_state,
			},
		)
		sb.Write(jsonString)
		sb.WriteString("\n")
		if p.transition_writer != nil {
			p.transition_writer.WriteString(sb.String())
		}
	}

	p.state = new_state
}
