package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/prometheus/common/log"
)

type Scraper interface {
	Ping() error
	Scrape() ([]byte, error)
}

type scraper struct {
	uri *url.URL
}

func (s *scraper) doRequest(uri string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).Dial,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%v/%v", s.uri, uri), nil)
	return client.Do(req)
}

func (s *scraper) Ping() error {
	response, err := s.doRequest("ping")
	if err != nil {
		log.Debugf("Problem connecting to Chronos: %v\n", err)
		return err
	}

	if response.StatusCode != 200 {
		log.Debugf("Problem reading Chronos ping response: %s\n", response.Status)
		return err
	}

	log.Debug("Connected to Chronos!")
	return nil
}

func (s *scraper) Scrape() ([]byte, error) {
	response, err := s.doRequest("metrics")
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, err
}
