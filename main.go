package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"time"
)

func buildHostSet(urls []string) (map[string]bool, error) {
	hosts := make(map[string]bool)

	for _, u := range urls {
		parsedURL, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		if parsedURL.Scheme != "https" {
			return nil, fmt.Errorf("upstream '%v' is not https", u)
		}
		if net.ParseIP(parsedURL.Host) == nil {
			hosts[parsedURL.Host+"."] = true
		}
	}

	return hosts, nil
}

var listenAddr string
var httpTimeout time.Duration
var fallback string
var verbose bool

func init() {
	pflag.StringVarP(&listenAddr, "listen", "l", ":10053", "address to listen on")
	pflag.DurationVarP(&httpTimeout, "timeout", "t", 5*time.Second, "HTTP timeout")
	pflag.StringVarP(&fallback, "fallback", "f", "1.1.1.1:53", "fallback DNS address for resolving upstreams")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
}

func main() {
	pflag.Parse()

	rootLogger := logrus.New()
	rootLogger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	if verbose {
		rootLogger.SetLevel(logrus.DebugLevel)
	}

	httpClient := http.Client{Timeout: httpTimeout}
	rand.Seed(time.Now().UTC().UnixNano())

	upstreams := flag.Args()

	if len(upstreams) == 0 {
		rootLogger.Fatal("Must include at least one upstream")
	}

	upstreamHosts, err := buildHostSet(upstreams)
	if err != nil {
		rootLogger.WithError(err).Fatal("Failed to parse upstreams")
	}

	proxy := func(w dns.ResponseWriter, r *dns.Msg) (*logrus.Entry, error) {
		questionName := r.Question[0].Name

		log := rootLogger.WithField("Id", r.Id).WithField("Domain", r.Question[0].Name)

		msg, err := r.Pack()
		if err != nil {
			return log, err
		}

		reader := bytes.NewReader(msg)

		if upstreamHosts[questionName] {
			log = log.WithField("upstream", fallback)

			conn, netErr := net.Dial("udp", fallback)
			defer conn.Close()
			if netErr != nil {
				return log, netErr
			}

			if _, copyErr := io.Copy(conn, reader); copyErr != nil {
				return log, copyErr
			}

			if _, copyErr := io.Copy(w, conn); copyErr != nil {
				return log, copyErr
			}

			return log, nil
		}

		dnsURL := upstreams[rand.Intn(len(upstreams))]
		log = log.WithField("upstream", dnsURL)

		resp, err := httpClient.Post(dnsURL, "application/dns-message", reader)
		if err != nil {
			return log, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return log, fmt.Errorf("received HTTP %v from %v", resp.StatusCode, dnsURL)
		}

		if _, err := io.Copy(w, resp.Body); err != nil {
			return log, err
		}

		return log, nil
	}

	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		defer w.Close()
		start := time.Now()
		if log, err := proxy(w, r); err != nil {
			log.WithError(err).Warnf("Failed to proxy in %v", time.Now().Sub(start).Round(time.Millisecond))
			reply := &dns.Msg{}
			reply.SetRcode(r, dns.RcodeServerFailure)
			if writeErr := w.WriteMsg(reply); writeErr != nil {
				log.WithError(writeErr).Warn("Error writing failure response")
			}
		} else {
			log.Debugf("Proxied in %v", time.Now().Sub(start).Round(time.Millisecond))
		}
	})

	server := dns.Server{
		Net:  "udp",
		Addr: listenAddr,
	}

	rootLogger.Infof("Proxying to %v URL(s) with a %v timeout", len(upstreams), httpClient.Timeout)
	rootLogger.Infof("Listening on %v", listenAddr)
	if err := server.ListenAndServe(); err != nil {
		rootLogger.Fatal(err)
	}
}
