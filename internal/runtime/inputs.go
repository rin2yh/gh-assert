package runtime

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/rhysd/actionlint"
)

const (
	workflowDispatch = "workflow_dispatch"
	inputsJSON       = "GH_ASSERT_INPUTS"
)

func isReusableWorkflow(contractPath string) (bool, error) {
	const suffix = "_assert.yml"
	if !strings.HasSuffix(contractPath, suffix) {
		return false, nil
	}
	path := strings.TrimSuffix(contractPath, suffix) + ".yml"
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	workflow, parseErrors := actionlint.Parse(data)
	if len(parseErrors) > 0 {
		first := parseErrors[0]
		return false, fmt.Errorf("%s:%d:%d: %s", path, first.Line, first.Column, first.Message)
	}
	_, reusable := workflow.FindWorkflowCallEvent()
	return reusable, nil
}

func loadEventInputs() (map[string]string, error) {
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

func loadForwardedInputs() (map[string]string, error) {
	var inputs map[string]any
	if err := decode([]byte(os.Getenv(inputsJSON)), &inputs); err != nil {
		return nil, fmt.Errorf("%s: inputs context is not valid JSON: %w", inputsJSON, err)
	}
	if inputs == nil {
		return nil, fmt.Errorf("%s: inputs context must be a JSON object", inputsJSON)
	}
	return stringify(inputs)
}

func parseInputs(data []byte) (map[string]string, error) {
	var payload struct {
		Inputs map[string]any `json:"inputs"`
	}
	if err := decode(data, &payload); err != nil {
		return nil, fmt.Errorf("event payload is not valid JSON: %w", err)
	}
	return stringify(payload.Inputs)
}

func decode(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func stringify(values map[string]any) (map[string]string, error) {
	inputs := make(map[string]string, len(values))
	for name, value := range values {
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
