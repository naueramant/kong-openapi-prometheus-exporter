package cmd

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"api-usage/pkg/kong"
	"api-usage/pkg/swagger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Start the prometheus metrics server",
	Run:   RunMetrics,
}

func init() {
	rootCmd.AddCommand(metricsCmd)
}

var (
	spec   *swagger.Specification
	config *Config

	prom            *prometheus.Registry
	httpReqsTotal   *prometheus.CounterVec
	httpReqDuration *prometheus.HistogramVec
)

func RunMetrics(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	config = loadConfig()

	// Load OpenAPI specification

	logrus.WithFields(logrus.Fields{
		"url": config.OpenAPI.URL,
	}).Info("Loading OpenAPI specification")

	if err := loadSpecification(ctx); err != nil {
		logrus.WithError(err).Fatal("Failed to load OpenAPI specification")
	}

	// Start auto reload job

	if config.OpenAPI.Reload != nil {
		logrus.WithFields(logrus.Fields{
			"interval": *config.OpenAPI.Reload,
		}).Info("OpenAPI specification auto reload enabled")

		go startReloadSpecificationJob(ctx)
	}

	// Initialize prometheus metrics

	initMetrics()

	// Register HTTP handlers

	http.Handle("/metrics", promhttp.HandlerFor(prom, promhttp.HandlerOpts{
		Registry: prom,
	}))

	http.Handle("/logs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Debug("Received log")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		log, err := kong.ParseLog(
			r.Body,
		)

		logrus.WithField("log", *log).Trace("raw log")

		if err != nil {
			logrus.WithError(err).Debug("Failed to parse log")
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		pathNode, ok := spec.MatchPath(log.Request.Method, log.Request.URI)
		if ok {
			recordMetrics(log, pathNode)
		}

		w.WriteHeader(http.StatusOK)
	}))

	// Start http server

	logrus.WithFields(logrus.Fields{
		"port": config.Prometheus.Port,
		"path": config.Prometheus.Path,
	}).Info("Starting prometheus server")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		logrus.WithError(err).Fatal("Failed to start prometheus server")
	}
}

func loadSpecification(ctx context.Context) error {
	specStartTime := time.Now()
	isReloading := spec != nil

	var err error

	if config.OpenAPI.URL != "" {
		spec, err = swagger.LoadURL(ctx, config.OpenAPI.URL)
	} else if config.OpenAPI.File != "" {
		spec, err = swagger.LoadFile(ctx, config.OpenAPI.File)
	}
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"duration": time.Since(specStartTime),
		"title":    spec.Meta.Title,
		"version":  spec.Meta.Version,
	}).Infof("OpenAPI specification %s", func() string {
		if isReloading {
			return "reloaded"
		}

		return "loaded"
	}())

	return nil
}

func startReloadSpecificationJob(ctx context.Context) {
	go func() {
		for {
			select {
			case <-time.After(*config.OpenAPI.Reload):
				loadSpecification(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func initMetrics() {
	// Register prometheus metrics

	promInstance := prometheus.NewRegistry()

	headerLabels := []string{}
	if config.Metrics.Headers != nil {
		for _, header := range *config.Metrics.Headers {
			headerLabels = append(headerLabels, headerNameToLabelName(header))
		}
	}

	// http_requests_total metric

	httpRequestsTotalLabels := []string{"host", "method", "status", "path"}
	httpRequestsTotalLabels = append(httpRequestsTotalLabels, headerLabels...)

	requestMetric := prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "kong_openapi_exporter",
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests",
	}, httpRequestsTotalLabels)

	promInstance.MustRegister(requestMetric)

	// http_request_duration_miliseconds

	httpRequestDurationLabels := []string{"host", "method", "path"}
	httpRequestDurationLabels = append(httpRequestDurationLabels, headerLabels...)

	latencyMetric := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: "kong_openapi_exporter",
		Name:      "http_request_duration_miliseconds",
		Help:      "HTTP request duration in milliseconds",
		Buckets:   []float64{25, 50, 80, 100, 250, 400, 700, 1000, 2000, 5000, 10000, 30000, 60000},
	}, httpRequestDurationLabels)

	promInstance.MustRegister(latencyMetric)

	// Assign metrics to global variables

	prom = promInstance
	httpReqsTotal = requestMetric
	httpReqDuration = latencyMetric
}

func recordMetrics(log *kong.Log, pathNode *swagger.Node) {
	// Match the path

	pathNode, ok := spec.MatchPath(log.Request.Method, log.Request.URI)
	if !ok {
		return
	}

	// http_requests_total labels

	httpReqsTotalLabels := prometheus.Labels{
		"host":   log.Request.Headers["host"],
		"method": log.Request.Method,
		"status": strconv.Itoa(log.Response.Status),
		"path":   pathNode.Path,
	}
	if config.Metrics.Headers != nil {
		for _, header := range *config.Metrics.Headers {
			httpReqsTotalLabels[headerNameToLabelName(header)] = log.Request.Headers[header]
		}
	}

	// http_request_duration_miliseconds labels

	httpReqDurationLabels := prometheus.Labels{
		"host":   log.Request.Headers["host"],
		"method": log.Request.Method,
		"path":   pathNode.Path,
	}
	if config.Metrics.Headers != nil {
		for _, header := range *config.Metrics.Headers {
			httpReqDurationLabels[headerNameToLabelName(header)] = log.Request.Headers[header]
		}
	}

	// Increment counters and observe histograms

	httpReqsTotal.With(httpReqsTotalLabels).Inc()
	httpReqDuration.With(httpReqDurationLabels).Observe(float64(log.Latencies.Request))
}

func headerNameToLabelName(header string) string {
	return strings.Replace(header, "-", "_", -1)
}
