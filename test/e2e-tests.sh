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

source "$(dirname "$0")/e2e-common.sh"

# Script entry point.
initialize $@ --skip-istio-addon

install_istio || fail_test "Istio installation failed"

echo ">> Build the binary kn-operator for the operator plugin"
go build -o kn-operator ./cmd/kn-operator.go || fail_test "Failed to build the binary of the operator plugin"

echo ">> Install the Knative Operator ${ALPHA_VERSION}"
./kn-operator install -n ${OPERATOR_NAMESPACE} -v ${ALPHA_VERSION} || fail_test "Failed to install Knative Operator ${ALPHA_VERSION}"

echo ">> Verify the installation of Knative Operator ${ALPHA_VERSION}"
go_test_e2e -tags=alpha -timeout=20m ./test/e2e || failed=1

echo ">> Upgrade to the latest version of Knative Operator"
./kn-operator install -n ${OPERATOR_NAMESPACE} -v ${LATEST_VERSION} || fail_test "Failed to upgrade to the latest Knative Operator"

echo ">> Verify the installation of Knative Operator of the latest version"
go_test_e2e -tags=beta -timeout=20m ./test/e2e || failed=1

echo ">> Install Knative Serving"
./kn-operator install -c serving -n ${SERVING_NAMESPACE} || fail_test "Failed to install Knative Serving"

echo ">> Configure the resource with Knative Serving"
./kn-operator configure resources -c serving -n ${SERVING_NAMESPACE} --deployName activator \
  --container activator --limitMemory 1001M --limitCPU 2048m --requestMemory 999M \
  --requestCPU 1024m || fail_test "Failed to configure Knative Serving"

echo ">> Verify the resource configuration of Knative Serving Custom resource"
go_test_e2e -tags=servingresourceconfig -timeout=20m ./test/e2e || failed=1

echo ">> Install Knative Eventing"
./kn-operator install -c eventing -n ${EVENTING_NAMESPACE} || fail_test "Failed to install Knative Eventing"

echo ">> Configure the resource with Knative Eventing"
./kn-operator configure resources -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --container eventing-controller --limitMemory 1001M --limitCPU 2048m --requestMemory 999M \
  --requestCPU 1024m || fail_test "Failed to configure Knative Eventing"

echo ">> Verify the resource configuration of Knative Eventing Custom resource"
go_test_e2e -tags=eventingresourceconfig -timeout=20m ./test/e2e || failed=1

echo ">> Remove Knative Operator"
./kn-operator uninstall -n ${OPERATOR_NAMESPACE} || fail_test "Failed to remove Knative Operator"

(( failed )) && fail_test

success
