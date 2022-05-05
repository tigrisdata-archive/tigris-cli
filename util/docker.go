package util

import (
	"encoding/json"
	"fmt"
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

// DockerShowProgress shows docker like progress output on terminal
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

			if err == io.EOF {
				return nil
			}
			return err
		}

		if strings.HasPrefix(event.Status, "Digest:") || strings.HasPrefix(event.Status, "Status:") {
			Stdout("%s\n", event.Status)
			continue
		}

		index, ok := layers[event.ID]

		if !ok {
			idxs = append(idxs, event.ID)
			index = len(idxs)
			layers[event.ID] = index
		}

		diff := index - curIdx

		if ok {
			if diff > 1 {
				c.moveDown(diff - 1)
			} else if diff < 1 {
				c.moveUp(-1*diff + 1)
			}
		} else if diff > 1 {
			c.moveDown(diff)
		}

		curIdx = index

		c.clearLine()

		if event.Status == "Pull complete" {
			fmt.Printf("%s: %s\n", event.ID, event.Status)
		} else {
			fmt.Printf("%s: %s %s\n", event.ID, event.Status, event.Progress)
		}
	}
}

type cursor struct{}

func (c *cursor) hide() {
	fmt.Printf("\033[?25l")
}

func (c *cursor) show() {
	fmt.Printf("\033[?25h")
}

func (c *cursor) moveUp(rows int) {
	fmt.Printf("\033[%dF", rows)
}

func (c *cursor) moveDown(rows int) {
	fmt.Printf("\033[%dE", rows)
}

func (c *cursor) clearLine() {
	fmt.Printf("\033[2K")
}
