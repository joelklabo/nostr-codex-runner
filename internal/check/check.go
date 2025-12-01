package check

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result represents a single dependency check outcome.
type Result struct {
	Name     string
	Type     string
	Status   string // OK|MISSING|WARN
	Details  string
	Optional bool
}

// Checker defines an interface for running checks.
type Checker interface {
	Check(dep DepInput) Result
}

// DepInput is a simplified view from config.Dep.
type DepInput struct {
	Name        string
	Type        string
	Version     string
	Optional    bool
	Description string
	Hint        string
}

// BinaryChecker checks for a binary on PATH and optional version substring.
type BinaryChecker struct{}

func (BinaryChecker) Check(dep DepInput) Result {
	res := Result{Name: dep.Name, Type: dep.Type, Status: "OK"}
	path, err := exec.LookPath(dep.Name)
	if err != nil {
		res.Status = missingStatus(dep.Optional)
		res.Details = fmt.Sprintf("not found in PATH (%s)", dep.Hint)
		return res
	}
	if dep.Version != "" {
		out, _ := exec.Command(path, "--version").CombinedOutput()
		if !strings.Contains(string(out), dep.Version) {
			res.Status = missingStatus(dep.Optional)
			res.Details = fmt.Sprintf("found %s but version mismatch (need %s)", strings.TrimSpace(string(out)), dep.Version)
			return res
		}
	}
	res.Details = path
	return res
}

func missingStatus(optional bool) string {
	if optional {
		return "WARN"
	}
	return "MISSING"
}
