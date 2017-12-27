package main

import (
	"flag"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var (
	listenAddress = flag.String(
		"web.listen-address", ":9044",
		"Address to listen on for web interface and telemetry.")

	metricsPath = flag.String(
		"web.telemetry-path", "/metrics",
		"Path under which to expose metrics.")

	chronosUri = flag.String(
		"chronos.uri", "http://chronos.mesos:4400",
		"URI of Chronos")
	chronosTimeout = flag.Duration(
		"chronos.timeout", 10*time.Second,
		"Timeout allowed for a chronos scrape")
	chronosVerifyTLS = flag.Bool(
		"chronos.verify-tls", true,
		"Verify the chronos.uri TLS certificate.  Insecure if disabled.")
	chronosAuthBearerToken = flag.String(
		"chronos.auth-bearer-token", "",
		"Send an Authorization Bearer header to chronos.url if given.  It's more secure to set this via the environment variable CHRONOS_AUTH_BEARER_TOKEN")
)

func main() {
	flag.Parse()
	uri, err := url.Parse(*chronosUri)
	if err != nil {
		log.Fatal(err)
	}

	if *chronosAuthBearerToken == "" {
		s, found := os.LookupEnv("CHRONOS_AUTH_BEARER_TOKEN")
		if found {
			log.Debugf("auth bearer was found in CHRONOS_AUTH_BEARER_TOKEN, using that")
			chronosAuthBearerToken = &s
		} else {
			chronosAuthBearerToken = nil
		}
	}

	scraper_instance := &scraper{
		auth_bearer_token: chronosAuthBearerToken,
		timeout:           *chronosTimeout,
		uri:               uri,
		verify_tls:        *chronosVerifyTLS,
	}

	for {
		err := scraper_instance.Ping()
		if err == nil {
			break
		}

		log.Debugf("Problem connecting to Chronos: %v", err)
		log.Infof("Couldn't connect to Chronos! Trying again in %v", chronosTimeout)
		time.Sleep(*chronosTimeout)
	}

	exporter := NewExporter(scraper_instance)
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
           <head><title>Chronos Exporter</title></head>
           <body>
           <h1>Chronos Exporter</h1>
           <p><a href='` + *metricsPath + `'>Metrics</a></p>
           </body>
           </html>`))
	})

	log.Info("Starting Server: ", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
