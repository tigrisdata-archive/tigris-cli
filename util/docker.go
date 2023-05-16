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
	"encoding/json"
	"errors"
	"io"
	"strings"
)

type pullEvent struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
	Progress       string `json:"progress,omitempty"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

// DockerShowProgress shows docker like progress output on terminal.
func DockerShowProgress(reader io.Reader) error {
	dec := json.NewDecoder(reader)

	layers := make(map[string]int, 0)
	idxs := make([]string, 0)
	curIdx := 0

	c := cursor{}
	c.hide()

	for {
		var event pullEvent
		if err := dec.Decode(&event); err != nil {
			c.show()

			if errors.Is(err, io.EOF) {
				return nil
			}

			return err
		}

		if strings.HasPrefix(event.Status, "Digest:") || strings.HasPrefix(event.Status, "Status:") {
			Stdoutf("%s\n", event.Status)

			continue
		}

		index, ok := layers[event.ID]

		if !ok {
			idxs = append(idxs, event.ID)
			index = len(idxs)
			layers[event.ID] = index
		}

		diff := index - curIdx

		c.moveCursor(ok, diff)

		curIdx = index

		c.clearLine()

		if event.Status == "Pull complete" {
			Stdoutf("%s: %s\n", event.ID, event.Status)
		} else {
			Stdoutf("%s: %s %s\n", event.ID, event.Status, event.Progress)
		}
	}
}

type cursor struct{}

func (*cursor) hide() {
	Stdoutf("\033[?25l")
}

func (*cursor) show() {
	Stdoutf("\033[?25h")
}

func (*cursor) moveUp(rows int) {
	Stdoutf("\033[%dF", rows)
}

func (*cursor) moveDown(rows int) {
	Stdoutf("\033[%dE", rows)
}

func (*cursor) clearLine() {
	Stdoutf("\033[2K")
}

func (c *cursor) moveCursor(ok bool, diff int) {
	if ok {
		if diff > 1 {
			c.moveDown(diff - 1)
		} else if diff < 1 {
			c.moveUp(-1*diff + 1)
		}
	} else if diff > 1 {
		c.moveDown(diff)
	}
}
