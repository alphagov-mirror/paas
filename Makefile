.DEFAULT_GOAL := help

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: spec compile_platform_tests lint_yaml lint_terraform lint_shellcheck lint_concourse lint_ruby lint_posix_newlines lint_symlinks ## Run linting tests

.PHONY: scripts_spec
scripts_spec:
	cd scripts &&\
		go get -d -t . &&\
		go test

.PHONY: tools_spec
tools_spec:
	cd tools/metrics &&\
		go test -v $(go list ./... | grep -v acceptance)
	cd tools/user_emails &&\
		go test -v ./...
	cd tools/user_management &&\
		bundle exec rspec --format documentation

.PHONY: concourse_spec
concourse_spec:
	cd concourse &&\
		bundle exec rspec
	cd concourse/scripts &&\
		go get -d -t . &&\
		go test
	cd concourse/scripts &&\
		bundle exec rspec

.PHONY: cloud_config_manifests_spec
cloud_config_manifests_spec:
	cd manifests/cloud-config &&\
		bundle exec rspec

.PHONY: cf_manifest_spec
cf_manifest_spec:
	cd manifests/cf-manifest &&\
		bundle exec rspec

.PHONY: shared_manifest_spec
shared_manifest_spec:
	cd manifests/shared &&\
		bundle exec rspec

.PHONY: bosh_manifest_spec
bosh_manifest_spec:
	cd manifests/bosh-manifest &&\
		bundle exec rspec

.PHONY: concourse_manifest_spec
concourse_manifest_spec:
	cd manifests/concourse-manifest &&\
		bundle exec rspec

.PHONY: prometheus_manifest_spec
prometheus_manifest_spec:
	cd manifests/prometheus &&\
		bundle exec rspec

.PHONY: manifests_spec
manifests_spec: cloud_config_manifests_spec cf_manifest_spec prometheus_manifest_spec shared_manifest_spec bosh_manifest_spec concourse_manifest_spec

.PHONY: terraform_spec
terraform_spec:
	cd terraform/scripts &&\
		go get -d -t . &&\
		go test
	cd terraform &&\
		bundle exec rspec

.PHONY: platform_tests_spec
platform_tests_spec:
	cd platform-tests &&\
		./run_tests.sh src/platform/availability/monitor/

.PHONY: config_spec
config_spec:
	cd config &&\
		bundle exec rspec

.PHONY: spec
spec: config_spec scripts_spec tools_spec concourse_spec manifests_spec terraform_spec platform_tests_spec

.PHONY: compile_platform_tests
compile_platform_tests:
	GOPATH="$$(pwd)/platform-tests" \
	go test -run ^$$ \
		platform/acceptance \
		platform/availability/api \
		platform/availability/app \
		platform/availability/helpers \
		platform/availability/monitor

.PHONY: lint_yaml
lint_yaml:
	find . -name '*.yml' -not -path '*/vendor/*' -not -path './manifests/prometheus/upstream/*' -not -path './manifests/cf-deployment/ci/template/*' | xargs yamllint -c yamllint.yml

.PHONY: lint_terraform
lint_terraform: dev ## Lint the terraform files.
	@scripts/lint_terraform.sh
	$(eval export TF_VAR_system_dns_zone_name=$SYSTEM_DNS_ZONE_NAME)
	$(eval export TF_VAR_apps_dns_zone_name=$APPS_DNS_ZONE_NAME)
	@terraform/scripts/lint.sh

.PHONY: lint_shellcheck
lint_shellcheck:
	find . -name '*.sh' -not -path './.git/*' -not -path '*/vendor/*' -not -path './platform-tests/pkg/*'  -not -path './manifests/cf-deployment/*' -not -path './manifests/prometheus/upstream/*' -not -path './manifests/bosh-manifest/upstream/*' | xargs shellcheck

.PHONY: lint_concourse
lint_concourse:
	cd .. && SHELLCHECK_OPTS="-e SC1091" python paas/concourse/scripts/pipecleaner.py --fatal-warnings paas/concourse/pipelines/*.yml

.PHONY: lint_ruby
lint_ruby:
	bundle exec govuk-lint-ruby

.PHONY: lint_posix_newlines
lint_posix_newlines:
	@# for some reason `git ls-files` is including 'manifests/cf-deployment' in its output...which is a directory
	git ls-files | grep -v -e vendor/ -e manifests/cf-deployment -e manifests/prometheus/upstream | xargs ./scripts/test_posix_newline.sh

.PHONY: lint_symlinks
lint_symlinks:
	# This mini-test tests that our script correctly identifies hanging symlinks
	@rm -f "$$TMPDIR/test-lint_symlinks"
	@ln -s /this/does/not/exist "$$TMPDIR/test-lint_symlinks"
	! echo "$$TMPDIR/test-lint_symlinks" | ./scripts/test_symlinks.sh 2>/dev/null # If <<this<< errors, the script is broken
	@rm "$$TMPDIR/test-lint_symlinks"
	# Successful end of mini-test
	find . -type l -not -path '*/vendor/*' \
	| grep -v $$(git submodule foreach 'echo -e ^./$$path' --quiet) \
	| ./scripts/test_symlinks.sh

.PHONY: pingdom
pingdom: ## Use custom Terraform provider to set up Pingdom check
	$(if ${ACTION},,$(error Must pass ACTION=<plan|apply|...>))
	@terraform/scripts/set-up-pingdom.sh ${ACTION}

.PHONY: logit-filters
logit-filters:
	mkdir -p config/logit/output
	docker run --rm -it \
		-v $(CURDIR):/mnt:ro \
		-v $(CURDIR)/config/logit/output:/output:rw \
		-w /mnt \
		jruby:9.1-alpine ./scripts/generate_logit_filters.sh $(LOGSEARCH_BOSHRELEASE_TAG) $(LOGSEARCH_FOR_CLOUDFOUNDRY_TAG)
	@echo "updated $(CURDIR)/config/logit/output/generated_logit_filters.conf"
