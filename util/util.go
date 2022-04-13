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

package util

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/tigrisdata/tigrisdb-cli/config"
)

var Version string
var DefaultTimeout = 5 * time.Second

func LogConfigure() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	//Colored output to terminal and just JSON output to pipe
	var output io.Writer = os.Stderr
	if fileInfo, _ := os.Stdout.Stat(); fileInfo.Mode()&os.ModeCharDevice != 0 {
		output = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	}
	lvl, err := zerolog.ParseLevel("trace")
	if err != nil {
		log.Error().Err(err).Msg("error parsing log level. defaulting to info level")
		lvl = zerolog.InfoLevel
	}
	log.Logger = zerolog.New(output).Level(lvl).With().Timestamp().CallerWithSkipFrameCount(2).Stack().Logger()
}

func GetContext(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := DefaultTimeout
	if config.DefaultConfig.Timeout != 0 {
		timeout = config.DefaultConfig.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	return ctx, cancel
}

func Stdout(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format, args...)
}
