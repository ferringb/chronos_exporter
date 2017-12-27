package main

import (
	"flag"
	"net/http"
	"net/url"
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
)

func main() {
	flag.Parse()
	uri, err := url.Parse(*chronosUri)
	if err != nil {
		log.Fatal(err)
	}

	scraper_instance := &scraper{uri, *chronosTimeout}

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
