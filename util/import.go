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

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

func cleanupNULLValuesLow(m map[string]any) {
	for k, v := range m {
		switch val := v.(type) {
		case map[string]any:
			cleanupNULLValuesLow(val)
		case []any:
			if len(val) == 0 {
				log.Debug().Str("name", k).Msg("removed empty array")
				delete(m, k)
			}
		case nil:
			log.Debug().Str("name", k).Msg("removed empty value")
			delete(m, k)
		}
	}
}

// CleanupNULLValues cleans up NULL values and empty arrays from the document.
func CleanupNULLValues(doc json.RawMessage) json.RawMessage {
	var m map[string]any

	err := jsoniter.Unmarshal(doc, &m)
	Fatal(err, "unmarshal doc for cleanup")

	cleanupNULLValuesLow(m)

	doc, err = jsoniter.Marshal(&m)
	Fatal(err, "marshal doc after cleanup")

	return doc
}
