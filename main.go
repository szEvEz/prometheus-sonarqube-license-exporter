package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

const namespace = "sonarqube_license"

var (
	tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
)

type sonarQubeCollector struct {
	licenseExpiresAt *prometheus.Desc
	maxLoc           *prometheus.Desc
	currentLoc       *prometheus.Desc
}

func newSonarQubeCollector() *sonarQubeCollector {
	return &sonarQubeCollector{
		licenseExpiresAt: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "expires_at"),
			"The date the SonarQube license expires",
			nil,
			nil,
		),
		maxLoc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "max_loc"),
			"The maximum lines of code of the license",
			nil,
			nil,
		),
		currentLoc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "current_loc"),
			"Current lines of code",
			nil,
			nil,
		),
	}
}

func (collector *sonarQubeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.licenseExpiresAt
	ch <- collector.maxLoc
	ch <- collector.currentLoc
}

type sonarQubeMetrics struct {
	LicenseExpiresAt string  `json:"expiresAt"`
	MaxLoc           float64 `json:"maxLoc"`
	CurrentLoc       float64 `json:"loc"`
}

func validateEnvVars() (string, string) {
	sonarQubeURL, set := os.LookupEnv("SONARQUBE_URL")
	if !set {
		log.Fatal("SONARQUBE_URL environment variable is not set")
	}

	sonarQubeToken, set := os.LookupEnv("SONARQUBE_TOKEN")
	if !set {
		log.Fatal("SONARQUBE_TOKEN environment variable is not set")
	}

	return sonarQubeURL, sonarQubeToken
}

func (collector *sonarQubeCollector) Collect(ch chan<- prometheus.Metric) {

	sonarQubeURL, sonarQubeToken := validateEnvVars()

	req, err := http.NewRequest("GET", sonarQubeURL+"/api/editions/show_license?internal=true", nil)
	if err != nil {
		log.Warn(err)
	}

	// basic auth for sonarqube with token is username as token, password empty
	req.SetBasicAuth(sonarQubeToken, "")
	resp, err := client.Do(req)
	if err != nil {
		log.Warn(err)
	} else if resp.StatusCode == 401 {
		log.Warn(resp.StatusCode, " Unauthorized")
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warn(err)
	}

	var t sonarQubeMetrics
	err = json.Unmarshal(body, &t)
	if err != nil {
		log.Warn("unmarshal error", err)
	}

	ti, err := time.Parse("2006-01-02", t.LicenseExpiresAt)
	if err != nil {
		log.Warn(err)
	}

	log.Info("Metrics retrieved")

	ch <- prometheus.MustNewConstMetric(collector.licenseExpiresAt, prometheus.GaugeValue, float64(ti.Unix()))
	ch <- prometheus.MustNewConstMetric(collector.maxLoc, prometheus.GaugeValue, t.MaxLoc)
	ch <- prometheus.MustNewConstMetric(collector.currentLoc, prometheus.GaugeValue, t.CurrentLoc)
}

func main() {

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	validateEnvVars()

	log.Info("Service started on localhost:9191/metrics")
	sonarQubeCollector := newSonarQubeCollector()

	reg := prometheus.NewRegistry()
	reg.MustRegister(sonarQubeCollector)

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":9191", nil))
}
