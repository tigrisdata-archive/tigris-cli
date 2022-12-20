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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var (
	DefaultConfig = Config{}

	DefaultURL = "api.preview.tigrisdata.cloud"
	Domain     = "tigrisdata.cloud"
)

type Log struct {
	Level string `json:"level" yaml:"level,omitempty"`
}

type Config struct {
	ClientID     string        `json:"client_id" yaml:"client_id,omitempty" mapstructure:"client_id"`
	ClientSecret string        `json:"client_secret" yaml:"client_secret,omitempty" mapstructure:"client_secret"`
	Token        string        `json:"token" yaml:"token,omitempty"`
	URL          string        `json:"url" yaml:"url,omitempty"`
	UseTLS       bool          `json:"use_tls" yaml:"use_tls,omitempty" mapstructure:"use_tls"`
	Timeout      time.Duration `json:"timeout" yaml:"timeout,omitempty"`
	Protocol     string        `json:"protocol" yaml:"protocol,omitempty"`
	Log          Log           `json:"log" yaml:"log,omitempty"`
	Project      string        `json:"project" yaml:"project,omitempty"`
}

var DefaultName = "tigris-cli"

var configPath = []string{
	"/etc/tigris/",
	"$HOME/.tigris",
	"./config/",
	"./",
}

var envPrefix = "tigris"

func Save(name string, config interface{}) error {
	var home string

	if runtime.GOOS == "windows" {
		home = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
	} else {
		home = os.Getenv("HOME")
	}

	// if home is not set write to current directory
	path := "."
	if home != "" {
		path = home
	}

	path += "/.tigris/"
	if err := os.MkdirAll(path, 0o700); err != nil {
		return err
	}

	file := path + name + ".yaml"
	if err := os.Rename(file, file+".bak"); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	b, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := os.WriteFile(file, b, 0o600); err != nil {
		return err
	}

	return nil
}

func Load(name string, config interface{}) {
	viper.SetConfigName(name + ".yaml")
	viper.SetConfigType("yaml")

	for _, v := range configPath {
		viper.AddConfigPath(v)
	}

	// This is needed to automatically bind environment variables to config struct
	// Viper will only bind environment variables to the keys it already knows about
	b, err := json.Marshal(config)
	if err != nil {
		e(err, "marshal config")
	}

	// This is needed to replace periods with underscores when mapping environment variables to multi-level config keys
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// The environment variables have a higher priority as compared to config values defined in the config file.
	// This allows us to override the config values using environment variables.
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()

	viper.SetConfigType("json")

	br := bytes.NewBuffer(b)
	if err = viper.MergeConfig(br); err != nil {
		e(err, "merge config")
	}

	viper.SetConfigType("yaml")

	err = viper.MergeInConfig()
	if err != nil {
		var ep viper.ConfigFileNotFoundError
		if !errors.As(err, &ep) {
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
