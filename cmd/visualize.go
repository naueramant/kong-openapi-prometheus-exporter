package cmd

import (
	"context"
	"time"

	"api-usage/pkg/swagger"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// visualizeCmd represents the visualize command
var visualizeCmd = &cobra.Command{
	Use:   "visualize",
	Short: "Generate a visualization of the API as a graph",
	Run:   RunVisualize,
}

func init() {
	rootCmd.AddCommand(visualizeCmd)

	visualizeCmd.PersistentFlags().String("out", "graph", "Output directory for the visualization files")

	visualizeCmd.PersistentFlags().String("url", "", "URL to the OpenAPI specification")
	visualizeCmd.PersistentFlags().String("file", "", "File path to the OpenAPI specification")
}

func RunVisualize(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	outDir, err := cmd.Flags().GetString("out")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get output directory")
	}
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get URL")
	}
	var filePath string
	if url == "" {
		filePath, err = cmd.Flags().GetString("file")
		if err != nil {
			logrus.WithError(err).Fatal("Failed to get file path")
		}
	}

	// logrus.WithFields(logrus.Fields{
	// 	"url": config.OpenAPI.URL,
	// }).Info("Loading specification")

	specStartTime := time.Now()

	var spec *swagger.Specification

	if url != "" {
		spec, err = swagger.LoadURL(ctx, url)
	} else if filePath != "" {
		spec, err = swagger.LoadFile(ctx, filePath)
	} else {
		logrus.Fatal("Either URL or file path must be provided")
	}

	if err != nil {
		logrus.WithError(err).Fatal("Failed to load specification")
	}

	logrus.WithFields(logrus.Fields{
		"duration": time.Since(specStartTime),
		"title":    spec.Meta.Title,
		"version":  spec.Meta.Version,
	}).Info("Specification loaded")

	logrus.WithField("out", outDir).Info("Writing visualization files")

	swagger.Visualize(spec, outDir)

	logrus.Info("Done")
}
