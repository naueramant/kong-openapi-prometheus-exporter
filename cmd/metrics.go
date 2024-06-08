package cmd

import (
	"context"
	"fmt"
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

var spec *swagger.Specification

func RunMetrics(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	logrus.WithFields(logrus.Fields{
		"url": config.OpenAPI.URL,
	}).Info("Loading OpenAPI specification")

	loadSpecifcation(ctx, config.OpenAPI.URL, false)

	if config.OpenAPI.Reload != nil {
		go startReloadSpecificationJob(ctx)
	}

	promInstance := prometheus.NewRegistry()

	defaultLabels := []string{"method", "status", "duration", "path"}
	extraLabels := []string{}
	for _, header := range *config.Metrics.Headers {
		extraLabels = append(extraLabels, headerNameToLabelName(header))
	}

	metricLabels := append(defaultLabels, extraLabels...)

	requestMetric := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_api_total",
		Help: "Total number of requests to the API",
	}, metricLabels)

	promInstance.MustRegister(requestMetric)

	http.Handle("/metrics", promhttp.HandlerFor(promInstance, promhttp.HandlerOpts{
		Registry: promInstance,
	}))

	http.Handle("/log", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

		log, err := kong.ParseLog(
			r.Body,
		)
		if err != nil {
			fmt.Println(err)

			w.WriteHeader(http.StatusBadRequest)

			return
		}

		ok, specPath := spec.MatchPath(log.Request.Method, log.Request.URI)
		if ok {
			requestMetric.With(logToLabels(log, *specPath)).Inc()
		}

		w.WriteHeader(http.StatusOK)
	}))

	logrus.WithFields(logrus.Fields{
		"port": config.Prometheus.Port,
		"path": config.Prometheus.Path,
	}).Info("Starting prometheus server")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		logrus.WithError(err).Fatal("Failed to start prometheus server")
	}
}

func loadSpecifcation(ctx context.Context, url string, reloaded bool) error {
	specStartTime := time.Now()

	var err error
	spec, err = swagger.LoadURL(ctx, url)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"duration":            time.Since(specStartTime),
		"title":               spec.Meta.Title,
		"version":             spec.Meta.Version,
		"base_path":           spec.Meta.BasePath,
		"number_of_endpoints": spec.Meta.NumberOfEndpoints,
	}).Infof("OpenAPI specification %s", func() string {
		if reloaded {
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
				loadSpecifcation(ctx, config.OpenAPI.URL, true)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func logToLabels(log *kong.Log, path string) prometheus.Labels {
	labels := prometheus.Labels{
		"method":   log.Request.Method,
		"status":   strconv.Itoa(log.Response.Status),
		"duration": strconv.Itoa(log.Latencies.Request),
		"path":     path,
	}

	for _, header := range *config.Metrics.Headers {
		labels[headerNameToLabelName(header)] = log.Request.Headers[header]
	}

	return labels
}

func headerNameToLabelName(header string) string {
	return strings.Replace(header, "-", "_", -1)
}
