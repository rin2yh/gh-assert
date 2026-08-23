package github

type Action struct {
	Inputs map[string]ActionInput `yaml:"inputs"`
	Runs   ActionRuns             `yaml:"runs"`
}

type ActionInput struct {
	Required bool `yaml:"required"`
}

type ActionRuns struct {
	Using string `yaml:"using"`
}
