package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/concourse/concourse/atc"
	"sigs.k8s.io/yaml"
)

type pipelineValidator struct {
	Filename string

	JobValidators      []jobValidator
	ResourceValidators []resourceValidator
	TaskValidators     []taskValidator

	ParseError                  error
	ConcourseValidationErrors   []error
	ConcourseValidationWarnings []error
	JobErrors                   []jobErrorCollection
	ResourceErrors              []resourceErrorCollection

	PipelineConfig atc.Config
}

func (pv *pipelineValidator) HasErrors() bool {
	if pv.ParseError != nil {
		return true
	}

	if len(pv.ConcourseValidationErrors) > 0 {
		return true
	}

	if len(pv.ConcourseValidationWarnings) > 0 {
		return true
	}

	for _, resourceErrors := range pv.ResourceErrors {
		for _, resourceResourceErrors := range resourceErrors.ResourceErrors {
			if len(resourceResourceErrors.ValidatorErrors) > 0 {
				return true
			}
		}
	}

	for _, jobErrors := range pv.JobErrors {
		for _, jobJobErrors := range jobErrors.JobErrors {
			if len(jobJobErrors.ValidatorErrors) > 0 {
				return true
			}
		}

		for _, resourceErrors := range jobErrors.ResourceErrors {
			for _, resourceResourceErrors := range resourceErrors.ResourceErrors {
				if len(resourceResourceErrors.ValidatorErrors) > 0 {
					return true
				}
			}
		}

		for _, taskErrors := range jobErrors.TaskErrors {
			for _, taskTaskErrors := range taskErrors.TaskErrors {
				if len(taskTaskErrors.ValidatorErrors) > 0 {
					return true
				}
			}
		}
	}

	return false
}

func (pv *pipelineValidator) FailureOutput() string {
	lines := []string{
		fmt.Sprintf("FILE %s", rColor(pv.Filename)),
	}

	if pv.ParseError != nil {
		lines = append(lines, indent(fmt.Sprintf("PARSE %s", pv.ParseError)))
		return strings.Join(lines, "\n")
	}

	if len(pv.ConcourseValidationErrors) > 0 {
		for _, err := range pv.ConcourseValidationErrors {
			lines = append(lines, indent(fmt.Sprintf("CONCOURSE %s", err)))
		}
		return strings.Join(lines, "\n")
	}

	if len(pv.ConcourseValidationWarnings) > 0 {
		for _, err := range pv.ConcourseValidationWarnings {
			lines = append(lines, indent(fmt.Sprintf("CONCOURSE %s", err)))
		}
	}

	for _, resourceErrors := range pv.ResourceErrors {
		failureOutput := resourceErrors.FailureOutput()
		if failureOutput != "" {
			lines = append(lines, indent(failureOutput))
		}
	}

	for _, jobErrors := range pv.JobErrors {
		failureOutput := jobErrors.FailureOutput()
		if failureOutput != "" {
			lines = append(lines, indent(failureOutput))
		}
	}

	return strings.Join(lines, "\n")
}

func (pv *pipelineValidator) SuccessOutput() string {
	return fmt.Sprintf("FILE %s", gColor(pv.Filename))
}

func (pv *pipelineValidator) Validate() {
	pv.ParseAndValidate()

	if pv.ParseError != nil || len(pv.ConcourseValidationErrors) > 0 {
		return
	}

	pv.ValidateResources()
	pv.ValidateJobs()
}

func (pv *pipelineValidator) ParseAndValidate() {
	var config atc.Config
	var err error

	contents, err := ioutil.ReadFile(pv.Filename)
	if err != nil {
		pv.ParseError = err
		return
	}

	// It is a common pattern to have a variable for trigger
	// we should set this to a boolean, so that we can parse as YAML
	contents = resourceTriggerRegexp.ReplaceAllLiteral(
		contents,
		[]byte("trigger: true"),
	)

	err = yaml.Unmarshal(contents, &config)
	if err != nil {
		pv.ParseError = err
	}

	pv.PipelineConfig = config
	validateWarnings, validateErrors := pv.PipelineConfig.Validate()

	for _, warning := range validateWarnings {
		pv.ConcourseValidationWarnings = append(
			pv.ConcourseValidationWarnings,
			fmt.Errorf("atc validate warning %s: %s", warning.Type, warning.Message),
		)
	}

	for _, err := range validateErrors {
		pv.ConcourseValidationErrors = append(
			pv.ConcourseValidationErrors,
			fmt.Errorf(err),
		)
	}
}

func (pv *pipelineValidator) ValidateResources() {
	errors := make([]resourceErrorCollection, 0)

	for _, resource := range pv.PipelineConfig.Resources {
		errColl := resourceErrorCollection{ResourceName: resource.Name}

		for _, validator := range pv.ResourceValidators {
			errColl.ResourceErrors = append(errColl.ResourceErrors,
				validatorErrorCollection{
					ValidatorName:   validator.Name,
					ValidatorErrors: validator.ValidateFn(resource),
				},
			)
		}

		errors = append(errors, errColl)
	}

	pv.ResourceErrors = errors
}

func (pv *pipelineValidator) ValidateJobs() {
	errors := make([]jobErrorCollection, 0)

	for _, job := range pv.PipelineConfig.Jobs {
		errColl := jobErrorCollection{JobName: job.Name}

		for _, validator := range pv.JobValidators {
			errColl.JobErrors = append(errColl.JobErrors, validatorErrorCollection{
				ValidatorName:   validator.Name,
				ValidatorErrors: validator.ValidateFn(job),
			})
		}

		for _, plan := range job.Plans() {
			if plan.TaskConfig == nil {
				continue
			}

			errColl.TaskErrors = append(errColl.TaskErrors, taskErrorCollection{
				TaskName:   plan.Name(),
				TaskErrors: pv.ValidateTask(*plan.TaskConfig, plan.Params),
			})
		}

		errors = append(errors, errColl)
	}

	pv.JobErrors = errors
}

func (pv *pipelineValidator) ValidateTask(
	taskConfig atc.TaskConfig,
	params atc.Params,
) []validatorErrorCollection {
	errCollections := make([]validatorErrorCollection, 0)

	for _, validator := range pv.TaskValidators {
		errCollections = append(errCollections, validatorErrorCollection{
			ValidatorName:   validator.Name,
			ValidatorErrors: validator.ValidateFn(taskConfig, params),
		})
	}

	return errCollections
}
