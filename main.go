package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/beevik/ntp"
	"github.com/macrat/ayd/lib-ayd"
)

var (
	version = "HEAD"
	commit  = "UNKNOWN"
)

func NormalizeTarget(u *ayd.URL) *ayd.URL {
	u2 := &ayd.URL{
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
		logger.Failure("invalid target URI: host name is required", nil)
		return
	}

	stime := time.Now()
	resp, err := ntp.Query(args.TargetURL.Opaque)
	latency := time.Now().Sub(stime)

	if err != nil {
		if e, ok := err.(*net.OpError); ok && e.Op == "read" {
			logger.WithTime(stime, latency).Failure(fmt.Sprintf("%s: connection refused", e.Addr), nil)
		} else {
			logger.WithTime(stime, latency).Failure(err.Error(), nil)
		}
	} else {
		logger.WithTime(stime, resp.RTT).Healthy("query succeeded", map[string]interface{}{
			"reference": resp.ReferenceID,
			"stratum": resp.Stratum,
			"root_delay": resp.RootDelay.Seconds()*1000,
			"offset": resp.ClockOffset.Seconds()*1000,
		})
	}
}
