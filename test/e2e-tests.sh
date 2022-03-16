#!/usr/bin/env bash

# Copyright 2022 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script runs the end-to-end tests against Knative Serving built from source.
# It is started by prow for each PR. For convenience, it can also be executed manually.

# If you already have a Knative cluster setup and kubectl pointing
# to it, call this script with the --run-tests arguments and it will use
# the cluster and run the tests.

# Calling this script without arguments will create a new cluster in
# project $PROJECT_ID, start knative in it, run the tests and delete the
# cluster.

export GO111MODULE=on

export OPERATOR_NAMESPACE="${OPERATOR_NAMESPACE:-operator-test}"
export SERVING_NAMESPACE="${SERVING_NAMESPACE:-serving-test}"
export EVENTING_NAMESPACE="${EVENTING_NAMESPACE:-eventing-test}"
export ALPHA_VERSION="${ALPHA_VERSION:-1.2.0}"
export LATEST_VERSION="${LATEST_VERSION:-latest}"

# source "$(dirname "$0")/e2e-common.sh"
source "$(dirname "${BASH_SOURCE[0]}")/../vendor/knative.dev/hack/e2e-tests.sh"

# Script entry point.
initialize $@ --skip-istio-addon

echo ">> Build the binary kn-operator for the operator plugin"
go build -o kn-operator ./cmd/kn-operator.go || fail_test "Failed to build the binary of the operator plugin"

echo ">> Install the Knative Operator ${ALPHA_VERSION}"
./kn-operator install -n ${OPERATOR_NAMESPACE} -v ${ALPHA_VERSION} || fail_test "Failed to install Knative Operator ${ALPHA_VERSION}"

echo ">> Upgrade to the latest version of Knative Operator"
./kn-operator install -n ${OPERATOR_NAMESPACE} -v ${LATEST_VERSION} || fail_test "Failed to upgrade to the latest Knative Operator"

echo ">> Verify the installationg of Knative Operator"
#TODO

echo ">> Install Knative Serving"
./kn-operator install -c eventing -n ${SERVING_NAMESPACE} || fail_test "Failed to install Knative Serving"

echo ">> Verify the installation of Knative Serving"
go_test_e2e -timeout=20m ./test/e2e || failed=1

echo ">> Install Knative Eventing"
./kn-operator install -c eventing -n ${EVENTING_NAMESPACE} || fail_test "Failed to install Knative Eventing"

echo ">> Verify the installation of Knative Eventing"
#TODO

echo ">> Remove Knative Serving"
./kn-operator uninstall -c serving -n ${SERVING_NAMESPACE} || fail_test "Failed to remove Knative Serving"

echo ">> Remove Knative Eventing"
./kn-operator uninstall -c eventing -n ${EVENTING_NAMESPACE} || fail_test "Failed to remove Knative Eventing"

echo ">> Remove Knative Operator"
./kn-operator uninstall -n ${OPERATOR_NAMESPACE} || fail_test "Failed to remove Knative Operator"

(( failed )) && fail_test

success
