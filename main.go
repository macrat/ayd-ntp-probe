package main

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/beevik/ntp"
	"github.com/macrat/ayd/lib-ayd"
)

var (
	version = "HEAD"
	commit  = "UNKNOWN"
)

func NormalizeTarget(u *url.URL) *url.URL {
	u2 := &url.URL{
		Scheme: "ntp",
		Opaque: u.Opaque,
	}

	if u2.Opaque == "" {
		u2.Opaque = u.Host
	}

	return u2
}

func main() {
	flag.Usage = func() {
		fmt.Println("NTP plugin for Ayd?")
		fmt.Println()
		fmt.Println("usage: ayd-ntp-probe TARGET_URI")
	}
	showVersion := flag.Bool("v", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("ayd-ntp-probe %s (%s)\n", version, commit)
		return
	}

	args, err := ayd.ParseProbePluginArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, "$ ayd-ntp-probe TARGET_URI")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	args.TargetURL = NormalizeTarget(args.TargetURL)
	logger := ayd.NewLogger(args.TargetURL)

	if args.TargetURL.Opaque == "" {
		logger.Failure("invalid target URI: host name is required")
		return
	}

	stime := time.Now()
	resp, err := ntp.Query(args.TargetURL.Opaque)
	latency := time.Now().Sub(stime)

	if err != nil {
		if e, ok := err.(*net.OpError); ok && e.Op == "read" {
			logger.WithTime(stime, latency).Failure(fmt.Sprintf("%s: connection refused", e.Addr))
		} else {
			logger.WithTime(stime, latency).Failure(err.Error())
		}
	} else {
		logger.WithTime(stime, resp.RTT).Healthy(fmt.Sprintf(
			"reference=%X stratum=%d rootdelay=%+f offset=%+f",
			resp.ReferenceID,
			resp.Stratum,
			resp.RootDelay.Seconds(),
			resp.ClockOffset.Seconds(),
		))
	}
}
