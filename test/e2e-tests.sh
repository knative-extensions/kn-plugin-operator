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
export TEST_KEY="${TEST_KEY:-test-key}"
export TEST_VALUE="${TEST_VALUE:-test-value}"
export TEST_KEY_ADDITIONAL="${TEST_KEY_ADDITIONAL:-test-key-additional}"
export TEST_VALUE_ADDITIONAL="${TEST_VALUE_ADDITIONAL:-test-value-additional}"
export REPLICA_NUM="${REPLICA_NUM:-4}"

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

echo ">> Configure the label for Knative Serving"
./kn-operator configure labels -c serving -n ${SERVING_NAMESPACE} --deployName activator \
  --key ${TEST_KEY} --value ${TEST_VALUE} --label || fail_test "Failed to configure Knative Serving"

./kn-operator configure labels -c serving -n ${SERVING_NAMESPACE} --deployName activator \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} --label || fail_test "Failed to configure Knative Serving"

echo ">> Configure the annotation for Knative Serving"
./kn-operator configure labels -c serving -n ${SERVING_NAMESPACE} --deployName activator \
  --key ${TEST_KEY} --value ${TEST_VALUE} --annotation || fail_test "Failed to configure Knative Serving"

./kn-operator configure labels -c serving -n ${SERVING_NAMESPACE} --deployName activator \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} --annotation || fail_test "Failed to configure Knative Serving"

echo ">> Configure the nodeSelector for Knative Serving"
./kn-operator configure labels -c serving -n ${SERVING_NAMESPACE} --deployName activator \
  --key ${TEST_KEY} --value ${TEST_VALUE} --nodeSelector || fail_test "Failed to configure Knative Serving"

./kn-operator configure labels -c serving -n ${SERVING_NAMESPACE} --deployName activator \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} --nodeSelector || fail_test "Failed to configure Knative Serving"

echo ">> Verify the label configuration of Knative Serving"
go_test_e2e -tags=servinglabelconfig -timeout=20m ./test/e2e || failed=1

echo ">> Configure the resource with Knative Serving"
./kn-operator configure resources -c serving -n ${SERVING_NAMESPACE} --deployName activator \
  --container activator --limitMemory 1001M --limitCPU 2048m --requestMemory 999M \
  --requestCPU 1024m || fail_test "Failed to configure Knative Serving"

echo ">> Verify the resource configuration of Knative Serving Custom resource"
go_test_e2e -tags=servingresourceconfig -timeout=20m ./test/e2e || failed=1

echo ">> Configure the ConfigMaps for Knative Serving"
./kn-operator configure configmaps -c serving -n ${SERVING_NAMESPACE} --cmName config-network \
  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure the ConfigMap for Knative Serving"

./kn-operator configure configmaps -c serving -n ${SERVING_NAMESPACE} --cmName config-network \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure the ConfigMap for Knative Serving"

./kn-operator configure configmaps -c serving -n ${SERVING_NAMESPACE} --cmName config-deployment \
  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure the ConfigMap for Knative Serving"

echo ">> Verify the configuration for config maps in Knative Serving"
go_test_e2e -tags=servingconfigmap -timeout=20m ./test/e2e || failed=1

echo ">> Configure the number of replicas for Knative Serving"
./kn-operator configure replicas -c serving -n ${SERVING_NAMESPACE} --deployName controller \
  --replicas ${REPLICA_NUM} || fail_test "Failed to configure the number of replias for Knative Serving"

./kn-operator configure replicas -c serving -n ${SERVING_NAMESPACE} --replicas ${REPLICA_NUM} || fail_test "Failed to configure the number of replias for Knative Serving"

echo ">> Verify the number of replicas for Knative Serving"
go_test_e2e -tags=servingha -timeout=20m ./test/e2e || failed=1

echo ">> Install Knative Eventing"
./kn-operator install -c eventing -n ${EVENTING_NAMESPACE} || fail_test "Failed to install Knative Eventing"

echo ">> Configure the resource with Knative Eventing"
./kn-operator configure resources -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --container eventing-controller --limitMemory 1001M --limitCPU 2048m --requestMemory 999M \
  --requestCPU 1024m || fail_test "Failed to configure Knative Eventing"

echo ">> Verify the resource configuration of Knative Eventing Custom resource"
go_test_e2e -tags=eventingresourceconfig -timeout=20m ./test/e2e || failed=1

echo ">> Configure the label for Knative Eventing"
./kn-operator configure labels -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY} --value ${TEST_VALUE} --label || fail_test "Failed to configure Knative Eventing"

./kn-operator configure labels -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} --label || fail_test "Failed to configure Knative Eventing"

echo ">> Configure the annotation for Knative Eventing"
./kn-operator configure labels -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY} --value ${TEST_VALUE} --annotation || fail_test "Failed to configure Knative Eventing"

./kn-operator configure labels -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} --annotation || fail_test "Failed to configure Knative Eventing"

echo ">> Configure the nodeSelector for Knative Eventing"
./kn-operator configure labels -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY} --value ${TEST_VALUE} --nodeSelector || fail_test "Failed to configure Knative Eventing"

./kn-operator configure labels -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} --nodeSelector || fail_test "Failed to configure Knative Eventing"

echo ">> Verify the label configuration of Knative Eventing"
go_test_e2e -tags=eventinglabelconfig -timeout=20m ./test/e2e || failed=1

echo ">> Configure the ConfigMaps for Knative Eventing"
./kn-operator configure configmaps -c eventing -n ${EVENTING_NAMESPACE} --cmName config-features \
  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure the ConfigMap for Knative Eventing"

./kn-operator configure configmaps -c eventing -n ${EVENTING_NAMESPACE} --cmName config-features \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure the ConfigMap for Knative Eventing"

./kn-operator configure configmaps -c eventing -n ${EVENTING_NAMESPACE} --cmName config-tracing \
  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure the ConfigMap for Knative Eventing"

echo ">> Verify the configuration for config maps in Knative Eventing"
go_test_e2e -tags=eventingconfigmap -timeout=20m ./test/e2e || failed=1

echo ">> Configure the number of replicas for Knative Eventing"
./kn-operator configure replicas -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --replicas ${REPLICA_NUM} || fail_test "Failed to configure the number of replias for Knative Eventing"

./kn-operator configure replicas -c eventing -n ${EVENTING_NAMESPACE} --replicas ${REPLICA_NUM} || fail_test "Failed to configure the number of replias for Knative Eventing"

echo ">> Verify the number of replicas for Knative Eventing"
go_test_e2e -tags=eventingha -timeout=20m ./test/e2e || failed=1

echo ">> Remove Knative Operator"
./kn-operator uninstall -n ${OPERATOR_NAMESPACE} || fail_test "Failed to remove Knative Operator"

(( failed )) && fail_test

success
