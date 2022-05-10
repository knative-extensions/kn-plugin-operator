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
export NIGHTLY_VERSION="${NIGHTLY_VERSION:-nightly}"
export TEST_KEY="${TEST_KEY:-test-key}"
export TEST_VALUE="${TEST_VALUE:-test-value}"
export TEST_KEY_ADDITIONAL="${TEST_KEY_ADDITIONAL:-test-key-additional}"
export TEST_VALUE_ADDITIONAL="${TEST_VALUE_ADDITIONAL:-test-value-additional}"
export REPLICA_NUM="${REPLICA_NUM:-4}"
export TOLERATION_KEY="${TOLERATION_KEY:-toleration-key}"
export OPERATION="${OPERATION:-Exists}"
export EFFECT="${EFFECT:-NoSchedule}"
export ADDITIONAL_TOLERATION_KEY="${ADDITIONAL_TOLERATION_KEY:-additional-toleration-key}"
export ADDITIONAL_OPERATION="${ADDITIONAL_OPERATION:-Equal}"
export ADDITIONAL_TOLERATION_VALUE="${ADDITIONAL_TOLERATION_VALUE:-additional-toleration-value}"
export ADDITIONAL_EFFECT="${ADDITIONAL_EFFECT:-NoExecute}"
export IMAGE_URL="${IMAGE_URL:-gcr.io/knative-releases/knative.dev/eventing/cmd/controller:latest}"
export IMAGE_KEY="${IMAGE_KEY:-eventing-controller}"
export DEFAULT_SERVING_IMAGE_URL="${DEFAULT_SERVING_IMAGE_URL:-"gcr.io/knative-releases/knative.dev/serving/cmd/\$\{NAME\}:latest"}"
export DEFAULT_EVENTING_IMAGE_URL="${DEFAULT_EVENTING_IMAGE_URL:-"gcr.io/knative-releases/knative.dev/eventing/cmd/\$\{NAME\}:latest"}"
export SERVING_IMAGE_KEY="${SERVING_IMAGE_KEY:-controller}"
export SERVING_IMAGE_URL="${SERVING_IMAGE_URL:-gcr.io/knative-releases/knative.dev/serving/cmd/controller:latest}"

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
./kn-operator install -n ${OPERATOR_NAMESPACE} -v ${NIGHTLY_VERSION} || fail_test "Failed to upgrade to the nightly built Knative Operator"

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

echo ">> Configure the label for Knative Serving's service"
./kn-operator configure labels -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
  --key ${TEST_KEY} --value ${TEST_VALUE} --label || fail_test "Failed to configure Knative Serving's service"

./kn-operator configure labels -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} --label || fail_test "Failed to configure Knative Serving's service"

echo ">> Configure the annotation for Knative Serving's service"
./kn-operator configure labels -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
  --key ${TEST_KEY} --value ${TEST_VALUE} --annotation || fail_test "Failed to configure Knative Serving's service"

./kn-operator configure labels -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} --annotation || fail_test "Failed to configure Knative Serving's service"

echo ">> Verify the label configuration of Knative Serving's service"
go_test_e2e -tags=servingservicelabelconfig -timeout=20m ./test/e2e || failed=1

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

echo ">> Configure the tolerations for Knative Serving"
./kn-operator configure tolerations -c serving -n ${SERVING_NAMESPACE} --deployName autoscaler \
  --key ${TOLERATION_KEY} --operator ${OPERATION} --effect ${EFFECT} || fail_test "Failed to configure the tolerations for Knative Serving"

./kn-operator configure tolerations -c serving -n ${SERVING_NAMESPACE} --deployName autoscaler \
  --key ${ADDITIONAL_TOLERATION_KEY} --operator ${ADDITIONAL_OPERATION} --value ${ADDITIONAL_TOLERATION_VALUE} --effect ${ADDITIONAL_EFFECT} || fail_test "Failed to configure the tolerations for Knative Serving"

echo ">> Verify the tolerations for Knative Serving"
go_test_e2e -tags=servingtolerations -timeout=20m ./test/e2e || failed=1

echo ">> Configure the image of the deployment for Knative Seving"
./kn-operator configure images -c serving -n ${SERVING_NAMESPACE} --deployName controller \
  --imageKey ${SERVING_IMAGE_KEY} --imageURL ${SERVING_IMAGE_URL} || fail_test "Failed to configure the image of the deployment for Knative Serving"

echo ">> Configure the image of all deployments for Knative Seving"
./kn-operator configure images -c serving -n ${SERVING_NAMESPACE} \
  --imageKey default --imageURL ${DEFAULT_SERVING_IMAGE_URL} || fail_test "Failed to configure the image of all deployments for Knative Serving"

echo ">> Verify the image configuration for Knative Seving"
go_test_e2e -tags=servingimage -timeout=20m ./test/e2e || failed=1

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

echo ">> Configure the label for Knative Eventing's service"
./kn-operator configure labels -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY} --value ${TEST_VALUE} --label || fail_test "Failed to configure Knative Eventing's service"

./kn-operator configure labels -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} --label || fail_test "Failed to configure Knative Eventing's service"

echo ">> Configure the annotation for Knative Eventing's service"
./kn-operator configure labels -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY} --value ${TEST_VALUE} --annotation || fail_test "Failed to configure Knative Eventing's service"

./kn-operator configure labels -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} --annotation || fail_test "Failed to configure Knative Eventing's service"

echo ">> Verify the label configuration of Knative Eventing's service"
go_test_e2e -tags=eventingservicelabelconfig -timeout=20m ./test/e2e || failed=1

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

echo ">> Configure the tolerations for Knative Eventing"
./kn-operator configure tolerations -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-webhook \
  --key ${TOLERATION_KEY} --operator ${OPERATION} --effect ${EFFECT} || fail_test "Failed to configure the tolerations for Knative Eventing"

./kn-operator configure tolerations -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-webhook \
  --key ${ADDITIONAL_TOLERATION_KEY} --operator ${ADDITIONAL_OPERATION} --value ${ADDITIONAL_TOLERATION_VALUE} --effect ${ADDITIONAL_EFFECT} || fail_test "Failed to configure the tolerations for Knative Eventing"

echo ">> Verify the tolerations for Knative Eventing"
go_test_e2e -tags=eventingtolerations -timeout=20m ./test/e2e || failed=1

echo ">> Configure the image of the deployment for Knative Eventing"
./kn-operator configure images -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --imageKey ${IMAGE_KEY} --imageURL ${IMAGE_URL} || fail_test "Failed to configure the image of the deployment for Knative Eventing"

echo ">> Configure the image of all deployments for Knative Eventing"
./kn-operator configure images -c eventing -n ${EVENTING_NAMESPACE} \
  --imageKey default --imageURL ${DEFAULT_EVENTING_IMAGE_URL} || fail_test "Failed to configure the image of all deployments for Knative Eventing"

echo ">> Verify the image configuration for Knative Eventing"
go_test_e2e -tags=eventingimage -timeout=20m ./test/e2e || failed=1

echo ">> Remove Knative Operator"
./kn-operator uninstall -n ${OPERATOR_NAMESPACE} || fail_test "Failed to remove Knative Operator"

(( failed )) && fail_test

success
