package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"
)

var Version = "v0.0.0"
var CommitHash = "dev"
var BuildTimestamp = "1970-01-01T00:00:00"
var Builder = "go version go1.xx.y os/platform"

type Options struct {
	quiet      *bool
	privileged *bool
	system     *bool
	log        *string
}

func main() {
	options := Options{}
	options.privileged = flag.Bool("privileged", false, "switch to privileged mode (default if run as root or on windows; ineffective with -s)")
	options.system = flag.Bool("s", false, "uses system's ping")
	options.quiet = flag.Bool("q", false, "quiet mode, disable live update")
	options.log = flag.String("log", "", "transition log `filename`")
	flag.Usage = usage
	flag.Parse()
	hosts := flag.Args()

	if len(hosts) == 0 {
		fmt.Println("no host provided")
		return
	}

	quitSig := make(chan bool)
	quitFlag := false

	transition_writer := &TransitionWriter{}
	if *options.log != "" {
		transition_writer.Init(*options.log, &quitFlag)
		defer transition_writer.Close()
	}

	wh := &WrapperHolder{}
	wh.InitHosts(hosts, options, transition_writer)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		wh.Stop()
		quitFlag = true
		quitSig <- true
	}()

	wh.Start()

	if !*options.quiet {
		display := NewDisplay(wh)
		display.Start()

		for !quitFlag {
			display.Update()
			time.Sleep(100 * time.Millisecond)
		}

		display.Stop()
	} else {
		fmt.Print(VersionString())
		for !quitFlag {
			wh.CalcStats(2 * 1e9)
			time.Sleep(100 * time.Millisecond)
		}
	}

	<-quitSig

}

func VersionString() string {
	return fmt.Sprintf("multiping %v-%v\n", Version, CommitHash)
}

func VersionStringLong() string {
	return fmt.Sprintf("multiping %v-%v (build on %v using %v)\nhttps://github.com/babs/multiping\n\n", Version, CommitHash, BuildTimestamp, Builder)
}

func usage() {
	fmt.Print(VersionStringLong())
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Println(`  host [hosts...]

Hosts can have the following form:
- hostname or ip or ip://hostname => ping (implementation used depends on '-s' flag)
- tcp://hostname:port or tcp://[ipv6]:port => tcp probing
    While using ip addresses, tcp:// can take IPv4 or IPv6 (w/ brackets), tcp4:// can only take IPv4 and tcp6:// only IPv6 (w/ brackets)

Hint on address family can be provided with the following form:
- ip://hostname and tcp://hostname resolves as default
- ip4://hostname and tcp4://hostname resolves as IPv4
- ip6://hostname and tcp6://hostname resolves as IPv6

Notes about implementation: tcp implementation between probing (S/SA/R) and full handshake depends on the platform`)
}
