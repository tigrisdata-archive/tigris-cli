// Copyright 2022-2023 Tigris Data, Inc.
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

	Project string

	errUnableToReadProject = fmt.Errorf("please specify project name")
)

type Log struct {
	Level string `json:"level" yaml:"level,omitempty"`
}

type Config struct {
	ClientID     string `json:"client_id"     mapstructure:"client_id"     yaml:"client_id,omitempty"`
	ClientSecret string `json:"client_secret" mapstructure:"client_secret" yaml:"client_secret,omitempty"`
	Token        string `json:"token"         yaml:"token,omitempty"`
	URL          string `json:"url"           yaml:"url,omitempty"`
	Protocol     string `json:"protocol"      yaml:"protocol,omitempty"`
	Project      string `json:"project"       yaml:"project,omitempty"`
	Branch       string `json:"branch"        yaml:"branch,omitempty"`
	DataDir      string `json:"data_dir"      mapstructure:"data_dir"      yaml:"data_dir,omitempty"`

	Log          Log           `json:"log"            yaml:"log,omitempty"`
	Timeout      time.Duration `json:"timeout"        yaml:"timeout,omitempty"`
	UseTLS       bool          `json:"use_tls"        mapstructure:"use_tls"        yaml:"use_tls,omitempty"`
	SkipLocalTLS bool          `json:"skip_local_tls" mapstructure:"skip_local_tls" yaml:"skip_local_tls,omitempty"`
}

var DefaultName = "tigris-cli"

var configPath = []string{
	"/etc/tigris/",
	"$HOME/.tigris",
	"./config/",
	"./",
}

var envPrefix = "tigris"

func Save(name string, config any) error {
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

	return os.WriteFile(file, b, 0o600)
}

func Load(name string, config any) {
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
	_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1) //nolint:revive
}

func GetProjectName() string {
	// first user supplied flag
	// second env variable
	// third config file
	if Project == "" {
		Project = DefaultConfig.Project
		if Project == "" {
			e(errUnableToReadProject, "unable to read project")
		}
	}

	return Project
}
