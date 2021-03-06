package main_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/lager"
	. "github.com/alphagov/paas-cf/tools/metrics"
	"github.com/aws/aws-sdk-go/aws"
	awsec "github.com/aws/aws-sdk-go/service/elasticache"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/alphagov/paas-cf/tools/metrics/pkg/elasticache"
	"github.com/alphagov/paas-cf/tools/metrics/pkg/elasticache/fakes"
	m "github.com/alphagov/paas-cf/tools/metrics/pkg/metrics"
)

var _ = Describe("Elasticache Gauges", func() {

	var (
		logger             lager.Logger
		log                *gbytes.Buffer
		elasticacheAPI     *fakes.FakeElastiCacheAPI
		elasticacheService *elasticache.ElasticacheService

		cacheParameterGroups []*awsec.CacheParameterGroup

		describeCacheParameterGroupsPagesStub = func(
			input *awsec.DescribeCacheParameterGroupsInput,
			fn func(*awsec.DescribeCacheParameterGroupsOutput, bool) bool,
		) error {
			for i, cacheParameterGroup := range cacheParameterGroups {
				page := &awsec.DescribeCacheParameterGroupsOutput{
					CacheParameterGroups: []*awsec.CacheParameterGroup{cacheParameterGroup},
				}
				if !fn(page, i+1 >= len(cacheParameterGroups)) {
					break
				}
			}
			return nil
		}

		cacheClusters []*awsec.CacheCluster

		describeCacheClustersPagesStub = func(
			input *awsec.DescribeCacheClustersInput,
			fn func(*awsec.DescribeCacheClustersOutput, bool) bool,
		) error {
			for i, cacheCluster := range cacheClusters {
				page := &awsec.DescribeCacheClustersOutput{
					CacheClusters: []*awsec.CacheCluster{cacheCluster},
				}
				if !fn(page, i+1 >= len(cacheClusters)) {
					break
				}
			}
			return nil
		}
	)

	BeforeEach(func() {
		logger = lager.NewLogger("logger")
		log = gbytes.NewBuffer()
		logger.RegisterSink(lager.NewWriterSink(log, lager.INFO))
		elasticacheAPI = &fakes.FakeElastiCacheAPI{}
		elasticacheAPI.DescribeCacheParameterGroupsPagesStub = describeCacheParameterGroupsPagesStub
		elasticacheAPI.DescribeCacheClustersPagesStub = describeCacheClustersPagesStub
		elasticacheService = &elasticache.ElasticacheService{Client: elasticacheAPI}
	})

	It("returns zero if there are no clusters", func() {
		gauge := ElasticCacheInstancesGauge(logger, elasticacheService, 1*time.Second)
		defer gauge.Close()

		var metric m.Metric
		Eventually(func() string {
			var err error
			metric, err = gauge.ReadMetric()
			Expect(err).NotTo(HaveOccurred())
			return metric.Name
		}, 3*time.Second).Should(Equal("aws.elasticache.node.count"))

		Expect(metric.Value).To(Equal(float64(0)))
		Expect(metric.Kind).To(Equal(m.Gauge))
	})

	It("returns the number of nodes", func() {
		cacheClusters = []*awsec.CacheCluster{
			{
				NumCacheNodes: aws.Int64(2),
			},
			{
				NumCacheNodes: aws.Int64(1),
			},
		}

		gauge := ElasticCacheInstancesGauge(logger, elasticacheService, 1*time.Second)
		defer gauge.Close()

		var metric m.Metric
		Eventually(func() string {
			var err error
			metric, err = gauge.ReadMetric()
			Expect(err).NotTo(HaveOccurred())
			return metric.Name
		}, 3*time.Second).Should(Equal("aws.elasticache.node.count"))

		Expect(metric.Value).To(Equal(float64(3)))
		Expect(metric.Kind).To(Equal(m.Gauge))
	})

	It("handles AWS API errors when getting the number of nodes", func() {
		gauge := ElasticCacheInstancesGauge(logger, elasticacheService, 1*time.Second)
		defer gauge.Close()

		awsErr := errors.New("some error")
		elasticacheAPI.DescribeCacheClustersPagesStub = func(
			input *awsec.DescribeCacheClustersInput,
			fn func(*awsec.DescribeCacheClustersOutput, bool) bool,
		) error {
			return awsErr
		}

		Eventually(func() error {
			metric, err := gauge.ReadMetric()
			Expect(metric.Name).To(Equal(""))
			return err
		}, 3*time.Second).Should(MatchError(awsErr))
	})

	It("returns zero if there are no cache parameter groups", func() {
		gauge := ElasticCacheInstancesGauge(logger, elasticacheService, 1*time.Second)
		defer gauge.Close()

		var metric m.Metric
		Eventually(func() string {
			var err error
			metric, err = gauge.ReadMetric()
			Expect(err).NotTo(HaveOccurred())
			return metric.Name
		}, 3*time.Second).Should(Equal("aws.elasticache.cache_parameter_group.count"))

		Expect(metric.Value).To(Equal(float64(0)))
		Expect(metric.Kind).To(Equal(m.Gauge))
	})

	It("returns zero if there are only default cache parameter groups", func() {
		cacheParameterGroups = []*awsec.CacheParameterGroup{
			&awsec.CacheParameterGroup{
				CacheParameterGroupName: aws.String("default.redis3.2"),
			},
		}
		gauge := ElasticCacheInstancesGauge(logger, elasticacheService, 1*time.Second)
		defer gauge.Close()

		var metric m.Metric
		Eventually(func() string {
			var err error
			metric, err = gauge.ReadMetric()
			Expect(err).NotTo(HaveOccurred())
			return metric.Name
		}, 3*time.Second).Should(Equal("aws.elasticache.cache_parameter_group.count"))

		Expect(metric.Value).To(Equal(float64(0)))
		Expect(metric.Kind).To(Equal(m.Gauge))
	})

	It("returns the number of cache parameter groups exluding the default ones", func() {
		cacheParameterGroups = []*awsec.CacheParameterGroup{
			&awsec.CacheParameterGroup{
				CacheParameterGroupName: aws.String("default.redis3.2"),
			},
			&awsec.CacheParameterGroup{
				CacheParameterGroupName: aws.String("group-1"),
			},
			&awsec.CacheParameterGroup{
				CacheParameterGroupName: aws.String("group-1"),
			},
		}

		gauge := ElasticCacheInstancesGauge(logger, elasticacheService, 1*time.Second)
		defer gauge.Close()

		var metric m.Metric
		Eventually(func() string {
			var err error
			metric, err = gauge.ReadMetric()
			Expect(err).NotTo(HaveOccurred())
			return metric.Name
		}, 3*time.Second).Should(Equal("aws.elasticache.cache_parameter_group.count"))

		Expect(metric.Value).To(Equal(float64(2)))
		Expect(metric.Kind).To(Equal(m.Gauge))
	})

	It("handles AWS API errors when getting the number of cache parameter groups", func() {
		gauge := ElasticCacheInstancesGauge(logger, elasticacheService, 1*time.Second)
		defer gauge.Close()

		awsErr := errors.New("some error")
		elasticacheAPI.DescribeCacheParameterGroupsPagesStub = func(
			input *awsec.DescribeCacheParameterGroupsInput,
			fn func(*awsec.DescribeCacheParameterGroupsOutput, bool) bool,
		) error {
			return awsErr
		}

		Eventually(func() error {
			metric, err := gauge.ReadMetric()
			Expect(metric.Name).To(Equal(""))
			return err
		}, 3*time.Second).Should(MatchError(awsErr))
	})

})
