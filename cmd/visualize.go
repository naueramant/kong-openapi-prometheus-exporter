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
}

func RunVisualize(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	logrus.WithFields(logrus.Fields{
		"url": config.OpenAPI.URL,
	}).Info("Loading specification")

	specStartTime := time.Now()

	spec, err := swagger.LoadURL(ctx, config.OpenAPI.URL)
	if err != nil {
		panic(err)
	}

	logrus.WithFields(logrus.Fields{
		"duration":            time.Since(specStartTime),
		"title":               spec.Meta.Title,
		"version":             spec.Meta.Version,
		"base_path":           spec.Meta.BasePath,
		"number_of_endpoints": spec.Meta.NumberOfEndpoints,
	}).Info("Specification loaded")

	outDir, err := cmd.Flags().GetString("out")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get output directory")
	}

	logrus.WithField("out", outDir).Info("Writing visualization files")

	swagger.Visualize(spec, outDir)

	logrus.Info("Done")
}
