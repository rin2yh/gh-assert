package github

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
