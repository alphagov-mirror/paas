package factory_test

import (
	"github.com/concourse/atc"
	"github.com/concourse/atc/scheduler/factory"
	"github.com/concourse/atc/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Factory Try Step", func() {
	var (
		resourceTypes atc.ResourceTypes

		buildFactory        factory.BuildFactory
		actualPlanFactory   atc.PlanFactory
		expectedPlanFactory atc.PlanFactory
	)

	BeforeEach(func() {
		actualPlanFactory = atc.NewPlanFactory(123)
		expectedPlanFactory = atc.NewPlanFactory(123)
		buildFactory = factory.NewBuildFactory(42, actualPlanFactory)

		resourceTypes = atc.ResourceTypes{
			{
				Name:   "some-custom-resource",
				Type:   "docker-image",
				Source: atc.Source{"some": "custom-source"},
			},
		}
	})

	Context("when there is a task wrapped in a try", func() {
		It("builds correctly", func() {
			actual, err := buildFactory.Create(atc.JobConfig{
				Plan: atc.PlanSequence{
					{
						Try: &atc.PlanConfig{
							Task: "first task",
						},
					},
					{
						Task: "second task",
					},
				},
			}, nil, resourceTypes, nil)
			Expect(err).NotTo(HaveOccurred())

			expected := expectedPlanFactory.NewPlan(atc.DoPlan{
				expectedPlanFactory.NewPlan(atc.TryPlan{
					Step: expectedPlanFactory.NewPlan(atc.TaskPlan{
						Name:          "first task",
						PipelineID:    42,
						ResourceTypes: resourceTypes,
					}),
				}),
				expectedPlanFactory.NewPlan(atc.TaskPlan{
					Name:          "second task",
					PipelineID:    42,
					ResourceTypes: resourceTypes,
				}),
			})

			Expect(actual).To(testhelpers.MatchPlan(expected))
		})
	})

	Context("when the try also has a hook", func() {
		It("builds correctly", func() {
			actual, err := buildFactory.Create(atc.JobConfig{
				Plan: atc.PlanSequence{
					{
						Try: &atc.PlanConfig{
							Task: "first task",
						},
						Success: &atc.PlanConfig{
							Task: "second task",
						},
					},
				},
			}, nil, nil, nil)
			Expect(err).NotTo(HaveOccurred())

			expected := expectedPlanFactory.NewPlan(atc.OnSuccessPlan{
				Step: expectedPlanFactory.NewPlan(atc.TryPlan{
					Step: expectedPlanFactory.NewPlan(atc.TaskPlan{
						Name:       "first task",
						PipelineID: 42,
					}),
				}),
				Next: expectedPlanFactory.NewPlan(atc.TaskPlan{
					Name:       "second task",
					PipelineID: 42,
				}),
			})

			Expect(actual).To(testhelpers.MatchPlan(expected))
		})
	})
})
