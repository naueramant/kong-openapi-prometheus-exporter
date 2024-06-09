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

var config *Config

func RunMetrics(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	config = loadConfig()

	logrus.WithFields(logrus.Fields{
		"url": config.OpenAPI.URL,
	}).Info("Loading OpenAPI specification")

	if err := loadSpecification(ctx); err != nil {
		logrus.WithError(err).Fatal("Failed to load OpenAPI specification")
	}

	if config.OpenAPI.Reload != nil {
		go startReloadSpecificationJob(ctx)
	}

	promInstance := prometheus.NewRegistry()

	metricLabels := []string{"method", "status", "duration", "path"}

	if config.Metrics.IncludeOperationID {
		metricLabels = append(metricLabels, "operation_id")
	}

	headerLabels := []string{}
	for _, header := range *config.Metrics.Headers {
		headerLabels = append(headerLabels, headerNameToLabelName(header))
	}
	metricLabels = append(metricLabels, headerLabels...)

	requestMetric := prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "kong_openapi_exporter",
		Name:      "http_requests_total",
		Help:      "Total number of requests to the API",
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

		pathNode, ok := spec.MatchPath(log.Request.Method, log.Request.URI)
		if ok {
			requestMetric.With(logToLabels(log, pathNode)).Inc()
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

func logToLabels(log *kong.Log, pathNode *swagger.Node) prometheus.Labels {
	labels := prometheus.Labels{
		"method":   log.Request.Method,
		"status":   strconv.Itoa(log.Response.Status),
		"duration": strconv.Itoa(log.Latencies.Request),
		"path":     pathNode.Path,
	}

	if config.Metrics.IncludeOperationID {
		labels["operation_id"] = pathNode.OperationID
	}

	for _, header := range *config.Metrics.Headers {
		labels[headerNameToLabelName(header)] = log.Request.Headers[header]
	}

	return labels
}

func headerNameToLabelName(header string) string {
	return strings.Replace(header, "-", "_", -1)
}
