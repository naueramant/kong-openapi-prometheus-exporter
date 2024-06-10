package cmd

import (
	"os"
	"time"

	"github.com/creasty/defaults"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var cfgFile string

type Config struct {
	Log struct {
		Level  string `mapstructure:"level" default:"info" validate:"oneof=panic fatal error warn info debug trace"`
		Format string `mapstructure:"format" default:"json" validate:"oneof=text json"`
	} `mapstructure:"log"`
	OpenAPI struct {
		URL    string         `mapstructure:"url" validate:"required_without=File,omitempty,url"`
		File   string         `mapstructure:"file" validate:"required_without=URL,omitempty,filepath"`
		Reload *time.Duration `mapstructure:"reload,omitempty"`
	} `mapstructure:"openapi"`
	Prometheus struct {
		Path string `mapstructure:"path" default:"/metrics"`
		Port string `mapstructure:"port" default:"8080"`
	}
	Metrics struct {
		Headers *[]string `mapstructure:"headers,omitempty"`
	}
}

var rootCmd = &cobra.Command{
	Short: "Kong OpenAPI prometheus exporter",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.yaml", "config file")
}

func loadConfig() *Config {
	var config Config

	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err == nil {
		logrus.WithField("file", viper.ConfigFileUsed()).Info("Using config file")
	}

	if err := defaults.Set(&config); err != nil {
		logrus.WithError(err).Fatal("Failed to set defaults")
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to unmarshal config")
	}

	if err := validator.New().Struct(config); err != nil {
		logrus.WithError(err).Fatal("Failed to validate config")
	}

	setupLogger(&config)

	return &config
}

func setupLogger(config *Config) {
	level, err := logrus.ParseLevel(config.Log.Level)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to parse log level")
	}
	logrus.SetLevel(level)

	var format string
	if config.Log.Format == "json" {
		format = "json"
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else if config.Log.Format == "text" {
		format = "text"
		logrus.SetFormatter(&logrus.TextFormatter{})
	} else {
		logrus.WithField("format", config.Log.Format).Fatal("Unsupported log format")
	}

	logrus.WithFields(logrus.Fields{
		"log_level":  level,
		"log_format": format,
	}).Info("Logger configured")
}
