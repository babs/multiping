package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

type Display struct {
	pwh                 *WrapperHolder
	noheader            bool
	area                *pterm.AreaPrinter
	host_format_string  string
	longest_host_string int
}

func NewDisplay(pwh *WrapperHolder) *Display {
	return &Display{
		pwh: pwh,
	}
}

func (d *Display) SetNoHeader(v bool) {
	d.noheader = v
}

func (d *Display) Start() {
	d.area, _ = pterm.DefaultArea.Start()
	d.longest_host_string = 0
	for _, wrapper := range d.pwh.ping_wrappers {
		if len(wrapper.Host()) > d.longest_host_string {
			d.longest_host_string = len(wrapper.Host())
		}
	}

	d.host_format_string = "%-" + fmt.Sprintf("%v", d.longest_host_string+2) + "s"
}

func (d *Display) Stop() {
	d.area.Stop()
}

func (d *Display) Update() {
	var sb strings.Builder
	if !d.noheader {
		sb.WriteString(VersionString())
	}

	for _, wrapper := range d.pwh.ping_wrappers {
		sb.WriteString(fmt.Sprintf(d.host_format_string, wrapper.Host()))
		stats := wrapper.CalcStats(2 * 1e9)
		if stats.last_seen_nano > 2*1e9 {
			if stats.lastrecv == 0 {
				sb.WriteString(bold_red.Sprintf("❌ never had reply"))
			} else {
				sb.WriteString(bold_red.Sprintf("❌ last reply %s ago", time.Duration(stats.last_seen_nano).Round(time.Second)))
			}
		} else {
			sb.WriteString(bold_green.Sprintf("✅ %-8s", stats.lastrtt_as_string))
			if stats.last_loss_nano > 0 {
				last_log := fmt.Sprintf(
					" (last loss %s: %s ago for %s)",
					time.Unix(0, stats.last_loss_nano).Format("2006-01-02 15:04:05"),
					time.Duration(time.Now().UnixNano()-stats.last_loss_nano).Round(time.Second),
					time.Duration(stats.last_loss_duration).Round(time.Second/10),
				)
				if d.longest_host_string+12+len(last_log) >= pterm.GetTerminalWidth() {
					sb.WriteString(fmt.Sprintf("\n%"+fmt.Sprintf("%v", pterm.GetTerminalWidth())+"s", last_log))
				} else {
					sb.WriteString(last_log)
				}
			}
		}
		sb.WriteString("\n")
	}

	d.area.Update(sb.String())
}

var bold_red = pterm.NewStyle(pterm.FgRed, pterm.Bold)
var bold_green = pterm.NewStyle(pterm.FgGreen, pterm.Bold)
