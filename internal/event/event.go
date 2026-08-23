package event

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
)

const Dispatch = "workflow_dispatch"

func Name() string { return os.Getenv("GITHUB_EVENT_NAME") }

func Inputs() (map[string]string, error) {
	path := os.Getenv("GITHUB_EVENT_PATH")
	if path == "" {
		return nil, errors.New("GITHUB_EVENT_PATH is not set")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	inputs, err := parseInputs(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return inputs, nil
}

func parseInputs(data []byte) (map[string]string, error) {
	var payload struct {
		Inputs map[string]any `json:"inputs"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("event payload is not valid JSON: %w", err)
	}
	inputs := make(map[string]string, len(payload.Inputs))
	for name, value := range payload.Inputs {
		text, ok := scalar(value)
		if !ok {
			return nil, fmt.Errorf("input %s is not a scalar value", name)
		}
		inputs[name] = text
	}
	return inputs, nil
}

func scalar(value any) (string, bool) {
	switch v := value.(type) {
	case nil:
		return "", true
	case string:
		return v, true
	case bool:
		return strconv.FormatBool(v), true
	case json.Number:
		return v.String(), true
	}
	return "", false
}
