package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Workflow struct {
	Events map[string]WorkflowEvent
}

type WorkflowEvent struct {
	Inputs map[string]WorkflowInput
}

type WorkflowInput struct {
	Required bool   `yaml:"required"`
	Type     string `yaml:"type"`
}

func LoadWorkflow(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseWorkflow(path, data)
}

func ParseWorkflow(path string, data []byte) (*Workflow, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	workflow := &Workflow{Events: map[string]WorkflowEvent{}}
	if len(document.Content) == 0 {
		return workflow, nil
	}
	on := mappingValue(document.Content[0], "on")
	for name, node := range eventNodes(on) {
		event := WorkflowEvent{}
		inputs := mappingValue(node, "inputs")
		if inputs != nil {
			if err := inputs.Decode(&event.Inputs); err != nil {
				return nil, fmt.Errorf("%s: %s.inputs: %w", path, name, err)
			}
		}
		workflow.Events[name] = event
	}
	return workflow, nil
}

func eventNodes(on *yaml.Node) map[string]*yaml.Node {
	events := map[string]*yaml.Node{}
	if on == nil {
		return events
	}
	switch on.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(on.Content); i += 2 {
			events[on.Content[i].Value] = on.Content[i+1]
		}
	case yaml.SequenceNode:
		for _, event := range on.Content {
			events[event.Value] = nil
		}
	case yaml.ScalarNode:
		events[on.Value] = nil
	}
	return events
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}
