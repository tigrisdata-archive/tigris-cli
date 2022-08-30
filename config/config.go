// Copyright 2022 Tigris Data, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var DefaultConfig = Config{
	URL:      "localhost:8081",
	Timeout:  5 * time.Second,
	Protocol: "grpc",
	UseTLS:   false,
}

type Config struct {
	ApplicationID     string        `json:"application_id" yaml:"application_id" mapstructure:"application_id"`
	ApplicationSecret string        `json:"application_secret" yaml:"application_secret" mapstructure:"application_secret"`
	URL               string        `json:"url" yaml:"url"`
	UseTLS            bool          `json:"use_tls" yaml:"use_tls" mapstructure:"use_tls"`
	Timeout           time.Duration `json:"timeout" yaml:"timeout"`
	Protocol          string        `json:"protocol" yaml:"protocol"`
}

var configPath = []string{
	"/etc/tigris/",
	"$HOME/.tigris",
	"./config/",
	"./",
}

var envPrefix = "tigris"

func Load(name string, config interface{}) {
	viper.SetConfigName(name + ".yaml")
	viper.SetConfigType("yaml")

	for _, v := range configPath {
		viper.AddConfigPath(v)
	}

	// This is needed to automatically bind environment variables to config struct
	// Viper will only bind environment variables to the keys it already knows about
	b, err := yaml.Marshal(config)
	if err != nil {
		e(err, "marshal config")
	}

	// This is needed to replace periods with underscores when mapping environment variables to multi-level config keys
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// The environment variables have a higher priority as compared to config values defined in the config file.
	// This allows us to override the config values using environment variables.
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()

	br := bytes.NewBuffer(b)
	if err = viper.MergeConfig(br); err != nil {
		e(err, "merge config")
	}

	viper.SetConfigType("")

	err = viper.MergeInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			//log.Warn().Err(err).Msgf("config file not found")
		} else {
			e(err, "error reading config")
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		e(err, "error unmarshalling config")
	}
}

func e(err error, _ string) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}
