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

package main

import (
	"os"

	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/cmd"
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/util"
)

func skipClientInit(args []string) bool {
	skip := map[string]bool{
		"local":      true,
		"version":    true,
		"completion": true,
		"docs":       true,
		"config":     true,
	}

	return len(args) > 1 && skip[args[1]]
}

func main() {
	util.LogConfigure()

	config.Load("tigris", &config.DefaultConfig)

	util.DefaultTimeout = config.DefaultConfig.Timeout

	if !skipClientInit(os.Args) {
		if err := client.Init(config.DefaultConfig); err != nil {
			util.Error(err, "tigris client initialization failed")
		}
	}

	cmd.Execute()
}
