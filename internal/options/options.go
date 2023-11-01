package options

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var Version string = "v0.0.1"

type PushgatewayAuthOptions struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type PushgatewayOptions struct {
	URL                string                 `yaml:"url"`
	InsecureSkipVerify bool                   `yaml:"insecureSkipVerify"`
	CAFile             string                 `yaml:"caFile"`
	CertFile           string                 `yaml:"certFile"`
	KeyFile            string                 `yaml:"keyFile"`
	Auth               PushgatewayAuthOptions `yaml:"auth"`
}

type ExporterOptions struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	Interval int    `yaml:"interval"`
}

type LogOptions struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

type Options struct {
	Pushgateway PushgatewayOptions `yaml:"pushgateway"`
	Exporters   []ExporterOptions  `yaml:"exporters"`
	Log         LogOptions         `yaml:"log"`
}

func NewOptions() (opts Options) {
	optsSource := viper.AllSettings()
	err := createOptions(optsSource, &opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create options failed:", err)
		os.Exit(1)
	}
	return
}
