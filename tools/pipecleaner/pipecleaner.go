package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	allResourcesUsed bool
	rubocop          bool
	shellcheck       bool
	secrets          bool
)

func main() {
	flag.BoolVar(
		&allResourcesUsed, "all-resources-used", true,
		"Enable/disable all-resources-used validator",
	)

	flag.BoolVar(
		&rubocop, "rubocop", true,
		"Enable/disable rubocop validator",
	)

	flag.BoolVar(
		&shellcheck, "shellcheck", true,
		"Enable/disable shellcheck validator",
	)

	flag.BoolVar(
		&secrets, "secrets", true,
		"Enable/disable secrets validator",
	)
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("pipecleaner")
		fmt.Println()
		fmt.Println("pipecleaner is a tool to analyse concourse pipelines/tasks")
		fmt.Println()
		fmt.Println("usage: pipecleaner [flags] pipeline1.yml [pipelineN.yml...]")
		fmt.Println()
		fmt.Println("flags:")
		flag.PrintDefaults()
		os.Exit(2)
	}

	jobValidators := make([]jobValidator, 0)
	resourceValidators := make([]resourceValidator, 0)
	taskValidators := make([]taskValidator, 0)

	if allResourcesUsed {
		jobValidators = append(jobValidators, allResourcesUsedValidator)
	}

	if rubocop {
		taskValidators = append(taskValidators, rubocopValidator)
	}

	if shellcheck {
		taskValidators = append(taskValidators, shellcheckValidator)
	}

	if secrets {
		resourceValidators = append(resourceValidators, secretsResourceValidator)
		taskValidators = append(taskValidators, secretsTaskValidator)
	}

	encounteredErrors := false

	for fileIndex, fname := range args {
		validator := pipelineValidator{
			Filename: fname,

			JobValidators:      jobValidators,
			ResourceValidators: resourceValidators,
			TaskValidators:     taskValidators,
		}

		validator.Validate()

		if fileIndex > 0 {
			fmt.Println()
		}

		if validator.HasErrors() {
			encounteredErrors = true
			fmt.Println(validator.FailureOutput())
		} else {
			fmt.Println(validator.SuccessOutput())
		}
	}

	if encounteredErrors {
		os.Exit(10)
	}

	os.Exit(0)
}

func shouldVarBeInterpolated(varname string) bool {
	if strings.Contains(strings.ToUpper(varname), "KEY") ||
		strings.Contains(strings.ToUpper(varname), "SECRET") {
		return true
	}
	return false
}
