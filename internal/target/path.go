package target

import (
	"path/filepath"
	"strings"
)

const (
	workflowContractSuffix = "_assert.yml"
	actionContractName     = "action_assert.yml"
)

func WorkflowPath(contractPath string) (string, bool) {
	dir := filepath.Dir(contractPath)
	if !strings.HasSuffix(filepath.Base(contractPath), workflowContractSuffix) ||
		filepath.Base(dir) != "workflows" || filepath.Base(filepath.Dir(dir)) != ".github" {
		return "", false
	}
	return strings.TrimSuffix(contractPath, workflowContractSuffix) + ".yml", true
}

func ActionPath(contractPath string) (string, bool) {
	if filepath.Base(contractPath) != actionContractName {
		return "", false
	}
	for dir := filepath.Dir(contractPath); ; dir = filepath.Dir(dir) {
		if filepath.Base(dir) == "actions" && filepath.Base(filepath.Dir(dir)) == ".github" {
			return filepath.Join(filepath.Dir(contractPath), "action.yml"), true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
	}
}
