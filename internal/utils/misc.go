package utils

import (
	"encoding/json"
	"slices"
	"strings"
)

type JSONObject map[string]any

func (o JSONObject) Validate() error {
	_, err := json.Marshal(o)
	return err
}

func EqualCollapsingWhitespace(a string, b string) bool {
	return slices.Equal(strings.Fields(a), strings.Fields(b))
}
