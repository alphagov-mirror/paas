#!/bin/bash

set -eu

if  [ "${DISABLE_CF_ACCEPTANCE_TESTS:-}" = "true" ]; then
  echo "WARNING: The acceptance tests have been disabled. Unset DISABLE_CF_ACCEPTANCE_TESTS when uploading the pipelines to enable them. You can still hijack this container to run them manually, but you must update the admin user in ./test-config/config.json."
  exit 0
fi

SLEEPTIME=90
NODES=5
SKIP_REGEX='routing.API|allows\spreviously-blocked\sip|Adding\sa\swildcard\sroute\sto\sa\sdomain|forwards\sapp\smessages\sto\sregistered\ssyslog\sdrains|when\sapp\shas\smultiple\sports\smapped'
SLOW_SPEC_THRESHOLD=120

export CONFIG
CONFIG="$(pwd)/test-config/config.json"

./paas-cf/concourse/scripts/import_bosh_ca.sh

echo "Linking acceptance-tests directory inside $GOPATH"
CF_GOPATH=/go/src/github.com/cloudfoundry
mkdir -p $CF_GOPATH
ln -s "$(pwd)/cf-release/src/github.com/cloudfoundry/cf-acceptance-tests" "${CF_GOPATH}/cf-acceptance-tests"

echo "Linking test artifacts directory"
ln -s "$(pwd)/artifacts" /tmp/artifacts

echo "Sleeping for ${SLEEPTIME} seconds..."
for i in $(seq $SLEEPTIME 1); do echo -ne "$i"'\r'; sleep 1; done; echo

cd "${CF_GOPATH}/cf-acceptance-tests"

echo "Starting acceptace tests"
./bin/test \
  -keepGoing \
  -randomizeAllSpecs \
  -skipPackage=helpers \
  -skip=${SKIP_REGEX} \
  -slowSpecThreshold=${SLOW_SPEC_THRESHOLD} \
  -nodes="${NODES}"
