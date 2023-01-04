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

package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/template"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/tigrisdata/tigris-cli/config"
)

var (
	Version        string
	DefaultTimeout = 5 * time.Second

	Quiet bool
)

func IsTTY(f *os.File) bool {
	fileInfo, _ := f.Stat()

	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func LogConfigure(cfg *config.Log) {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Colored output to terminal and just JSON output to pipe
	var output io.Writer = os.Stderr

	if IsTTY(os.Stdout) {
		output = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	}

	level := cfg.Level
	if cfg.Level == "" {
		level = "disabled"
	}

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Error().Err(err).Msg("error parsing log level. defaulting to info level")

		lvl = zerolog.InfoLevel
	}

	log.Logger = zerolog.New(output).Level(lvl).With().Timestamp().CallerWithSkipFrameCount(2).Stack().Logger()
}

func GetTimeout() time.Duration {
	timeout := DefaultTimeout
	if config.DefaultConfig.Timeout != 0 {
		timeout = config.DefaultConfig.Timeout
	}

	return timeout
}

func GetContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, GetTimeout())
}

func Stdoutf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}

func Stderrf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}

func PrettyJSON(s any) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	Stdoutf("%s\n", string(b))

	return nil
}

func PrintError(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "%s\n", err.Error())
}

func Infof(format string, args ...interface{}) {
	if !Quiet {
		Stdoutf(format+"\n", args...)
	}
}

func Error(err error, msg string, args ...interface{}) error {
	log.Err(err).CallerSkipFrame(3).Msgf(msg, args...)

	if err == nil {
		return nil
	}

	return err
}

func Fatal(err error, msg string, args ...interface{}) {
	if err == nil {
		return
	}

	PrintError(err)

	_ = Error(err, msg, args...)

	os.Exit(1)
}

func ExecTemplate(w io.Writer, tmpl string, vars interface{}) {
	t, err := template.New("exec_template").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).Parse(tmpl)
	Fatal(err, "error parsing template")

	err = t.Execute(w, vars)
	Fatal(err, "execute template failed")
}

func Contains(l []string, s string) bool {
	for _, v := range l {
		if v == s {
			return true
		}
	}

	return false
}
