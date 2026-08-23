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

	"github.com/rin2yh/gh-assert/internal/github"
	"github.com/rin2yh/gh-assert/internal/model"
)

const (
	workflowDispatch = "workflow_dispatch"
	inputsJSON       = "GH_ASSERT_INPUTS"
)

func inputs(item model.ContractFile) (map[string]string, error) {
	if len(item.Contract.Inputs) == 0 {
		return nil, nil
	}
	event := os.Getenv("GITHUB_EVENT_NAME")
	if event == workflowDispatch {
		inputs, err := eventInputs()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", item.Path, err)
		}
		return inputs, nil
	}

	workflow, err := workflow(item.Path)
	if err != nil {
		return nil, err
	}
	if workflow != nil {
		if _, ok := workflow.Events["workflow_call"]; ok {
			if os.Getenv(inputsJSON) == "" {
				return nil, fmt.Errorf("%s: reusable workflow inputs must be passed with workflow-inputs: ${{ toJSON(inputs) }}", item.Path)
			}
			inputs, err := workflowInputs()
			if err != nil {
				return nil, fmt.Errorf("%s: %w", item.Path, err)
			}
			return inputs, nil
		}
	}
	if event == "" {
		return nil, fmt.Errorf("%s: an inputs contract requires a %s run: GITHUB_EVENT_NAME is not set", item.Path, workflowDispatch)
	}
	return nil, fmt.Errorf("%s: an inputs contract requires a %s run, but the current event is %s", item.Path, workflowDispatch, event)
}

func workflow(contractPath string) (*github.Workflow, error) {
	const suffix = "_assert.yml"
	if !strings.HasSuffix(contractPath, suffix) {
		return nil, nil
	}
	path := strings.TrimSuffix(contractPath, suffix) + ".yml"
	workflow, err := github.NewParser(path).Parse()
	if os.IsNotExist(err) {
		return nil, nil
	}
	return workflow, err
}

func eventInputs() (map[string]string, error) {
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

func workflowInputs() (map[string]string, error) {
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
