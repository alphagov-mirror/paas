#!/bin/bash
set -e

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

env=${DEPLOY_ENV-$1}

[[ -z "${env}" ]] && echo "Must provide environment name" && exit 100

cf_manifest_dir="${SCRIPT_DIR}/../../manifests/cf-manifest/deployments"
cf_release_version=$("${SCRIPT_DIR}"/val_from_yaml.rb releases.cf.version "${cf_manifest_dir}/000-base-cf-deployment.yml")
cf_graphite_version=$("${SCRIPT_DIR}"/val_from_yaml.rb releases.graphite.version "${cf_manifest_dir}/055-graphite.yml")
cf_grafana_version=$("${SCRIPT_DIR}"/val_from_yaml.rb releases.grafana.version "${cf_manifest_dir}/055-graphite.yml")

generate_vars_file() {
   set -u # Treat unset variables as an error when substituting
   cat <<EOF
---
aws_account: ${AWS_ACCOUNT:-dev}
deploy_env: ${env}
state_bucket: ${env}-state
pipeline_trigger_file: ${pipeline_name}.trigger
branch_name: ${BRANCH:-master}
aws_region: ${AWS_DEFAULT_REGION:-eu-west-1}
debug: ${DEBUG:-}
cf-release-version: v${cf_release_version}
cf_graphite_version: ${cf_graphite_version}
cf_grafana_version: ${cf_grafana_version}
EOF
}

generate_manifest_file() {
   # This exists because concourse does not support boolean value interpolation by design
   enable_auto_deploy=$([ "${ENABLE_AUTO_DEPLOY:-}" ] && echo "true" || echo "false")
   sed -e "s/{{auto_deploy}}/${enable_auto_deploy}/" \
       < "${SCRIPT_DIR}/../pipelines/${pipeline_name}.yml"
}

pipeline_name="create-bosh-cloudfoundry"
generate_vars_file > /dev/null # Check for missing vars
bash "${SCRIPT_DIR}/deploy-pipeline.sh" \
  "${env}" "${pipeline_name}" \
  <(generate_manifest_file) \
  <(generate_vars_file)

for component in cloudfoundry microbosh; do
  pipeline_name="destroy-${component}"
  generate_vars_file > /dev/null # Check for missing vars
  bash "${SCRIPT_DIR}/deploy-pipeline.sh" \
    "${env}" "${pipeline_name}" \
    <(generate_manifest_file) \
    <(generate_vars_file)
done

pipeline_name="autodelete-cloudfoundry"
if [ ! "${DISABLE_AUTODELETE:-}" ]; then
  bash "${SCRIPT_DIR}/deploy-pipeline.sh" \
	  "${env}" "${pipeline_name}" \
    "${SCRIPT_DIR}/../pipelines/${pipeline_name}.yml" \
    <(generate_vars_file)

  echo
  echo "WARNING: Pipeline to autodelete Cloud Foundry has been setup and enabled."
  echo "         To disable it, set DISABLE_AUTODELETE=1 or pause the pipeline."
else
  yes y | ${FLY_CMD:-fly} -t "${FLY_TARGET:-$env}" destroy-pipeline --pipeline "${pipeline_name}" || true

  echo
  echo "WARNING: Pipeline to autodelete Cloud Foundry has NOT been setup"
fi
