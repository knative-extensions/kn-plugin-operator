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

export ENV_NAME="${ENV_NAME:-test-name}"
export ENV_VALUE="${ENV_VALUE:-test-value}"
export ADDITIONAL_ENV_NAME="${ADDITIONAL_ENV_NAME:-additional-test-name}"
export ADDITIONAL_ENV_VALUE="${ADDITIONAL_ENV_VALUE:-additional-test-value}"

source "$(dirname "$0")/e2e-common.sh"

# Script entry point.
initialize $@ --skip-istio-addon

install_istio || fail_test "Istio installation failed"

echo ">> Build the binary kn-operator for the operator plugin"
go build -o kn-operator ./cmd/kn-operator.go || fail_test "Failed to build the binary of the operator plugin"

echo ">> Upgrade to the latest version of Knative Operator"
./kn-operator install -n ${OPERATOR_NAMESPACE} -v ${NIGHTLY_VERSION} || fail_test "Failed to upgrade to the nightly built Knative Operator"

echo ">> Verify the installation of Knative Operator of the latest version"
go_test_e2e -tags=beta -timeout=20m ./test/e2e || failed=1

echo ">> Install Knative Serving"
./kn-operator install -c serving -n ${SERVING_NAMESPACE} || fail_test "Failed to install Knative Serving"

#echo ">> Configure the label for Knative Serving"
#./kn-operator configure labels -c serving -n ${SERVING_NAMESPACE} --deployName activator \
#  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure Knative Serving"
#
#./kn-operator configure labels -c serving -n ${SERVING_NAMESPACE} --deployName activator \
#  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure Knative Serving"
#
#echo ">> Configure the annotation for Knative Serving"
#./kn-operator configure annotations -c serving -n ${SERVING_NAMESPACE} --deployName activator \
#  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure Knative Serving"
#
#./kn-operator configure annotations -c serving -n ${SERVING_NAMESPACE} --deployName activator \
#  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure Knative Serving"
#
#echo ">> Configure the nodeSelector for Knative Serving"
#./kn-operator configure nodeSelectors -c serving -n ${SERVING_NAMESPACE} --deployName activator \
#  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure Knative Serving"
#
#./kn-operator configure nodeSelectors -c serving -n ${SERVING_NAMESPACE} --deployName activator \
#  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure Knative Serving"
#
#echo ">> Verify the label configuration of Knative Serving"
#go_test_e2e -tags=servinglabelconfig -timeout=20m ./test/e2e || failed=1
#
#echo ">> Configure the label for Knative Serving's service"
#./kn-operator configure labels -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
#  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure Knative Serving's service"
#
#./kn-operator configure labels -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
#  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure Knative Serving's service"
#
#echo ">> Configure the annotation for Knative Serving's service"
#./kn-operator configure annotations -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
#  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure Knative Serving's service"
#
#./kn-operator configure annotations -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
#  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure Knative Serving's service"
#
#echo ">> Configure the selector for Knative Serving's service"
#./kn-operator configure selectors -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
#  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure Knative Serving's service"
#
#./kn-operator configure selectors -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
#  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure Knative Serving's service"
#
#echo ">> Verify the label configuration of Knative Serving's service"
#go_test_e2e -tags=servingservicelabelconfig -timeout=20m ./test/e2e || failed=1
#
#echo ">> Configure the resource with Knative Serving"
#./kn-operator configure resources -c serving -n ${SERVING_NAMESPACE} --deployName activator \
#  --container activator --limitMemory 1001M --limitCPU 2048m --requestMemory 999M \
#  --requestCPU 1024m || fail_test "Failed to configure Knative Serving"
#
#echo ">> Verify the resource configuration of Knative Serving Custom resource"
#go_test_e2e -tags=servingresourceconfig -timeout=20m ./test/e2e || failed=1
#
#echo ">> Configure the ConfigMaps for Knative Serving"
#./kn-operator configure configmaps -c serving -n ${SERVING_NAMESPACE} --cmName config-network \
#  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure the ConfigMap for Knative Serving"
#
#./kn-operator configure configmaps -c serving -n ${SERVING_NAMESPACE} --cmName config-network \
#  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure the ConfigMap for Knative Serving"
#
#./kn-operator configure configmaps -c serving -n ${SERVING_NAMESPACE} --cmName config-deployment \
#  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure the ConfigMap for Knative Serving"
#
#echo ">> Verify the configuration for config maps in Knative Serving"
#go_test_e2e -tags=servingconfigmap -timeout=20m ./test/e2e || failed=1
#
echo ">> Configure the number of replicas for Knative Serving"
./kn-operator configure replicas -c serving -n ${SERVING_NAMESPACE} --deployName controller \
  --replicas ${REPLICA_NUM} || fail_test "Failed to configure the number of replias for Knative Serving"

./kn-operator configure replicas -c serving -n ${SERVING_NAMESPACE} --replicas ${REPLICA_NUM} || fail_test "Failed to configure the number of replias for Knative Serving"

echo ">> Verify the number of replicas for Knative Serving"
go_test_e2e -tags=servingha -timeout=20m ./test/e2e || failed=1
#
#echo ">> Configure the tolerations for Knative Serving"
#./kn-operator configure tolerations -c serving -n ${SERVING_NAMESPACE} --deployName autoscaler \
#  --key ${TOLERATION_KEY} --operator ${OPERATION} --effect ${EFFECT} || fail_test "Failed to configure the tolerations for Knative Serving"
#
#./kn-operator configure tolerations -c serving -n ${SERVING_NAMESPACE} --deployName autoscaler \
#  --key ${ADDITIONAL_TOLERATION_KEY} --operator ${ADDITIONAL_OPERATION} --value ${ADDITIONAL_TOLERATION_VALUE} --effect ${ADDITIONAL_EFFECT} || fail_test "Failed to configure the tolerations for Knative Serving"
#
#echo ">> Verify the tolerations for Knative Serving"
#go_test_e2e -tags=servingtolerations -timeout=20m ./test/e2e || failed=1
#
#echo ">> Configure the image of the deployment for Knative Seving"
#./kn-operator configure images -c serving -n ${SERVING_NAMESPACE} --deployName controller \
#  --imageKey ${SERVING_IMAGE_KEY} --imageURL ${SERVING_IMAGE_URL} || fail_test "Failed to configure the image of the deployment for Knative Serving"
#
#echo ">> Configure the image of all deployments for Knative Seving"
#./kn-operator configure images -c serving -n ${SERVING_NAMESPACE} \
#  --imageKey default --imageURL ${DEFAULT_SERVING_IMAGE_URL} || fail_test "Failed to configure the image of all deployments for Knative Serving"
#
#echo ">> Verify the image configuration for Knative Seving"
#go_test_e2e -tags=servingimage -timeout=20m ./test/e2e || failed=1
#
#echo ">> Configure the environment variables for the container in the deployment of Knative Seving"
#./kn-operator configure envvars -c serving -n ${SERVING_NAMESPACE} --deployName controller --container controller \
#  --name ${ENV_NAME} --value ${ENV_VALUE} || fail_test "Failed to configure the env var for Knative Serving"
#
#./kn-operator configure envvars -c serving -n ${SERVING_NAMESPACE} --deployName controller --container controller \
#  --name ${ADDITIONAL_ENV_NAME} --value ${ADDITIONAL_ENV_VALUE} || fail_test "Failed to configure the env var for Knative Serving"
#
#./kn-operator configure envvars -c serving -n ${SERVING_NAMESPACE} --deployName activator --container activator \
#  --name ${ENV_NAME} --value ${ENV_VALUE} || fail_test "Failed to configure the env var for Knative Serving"
#
#echo ">> Verify the env var configuration for Knative Serving"
#go_test_e2e -tags=servingenvvar -timeout=20m ./test/e2e || failed=1
#
#echo ">> Remove the resource configuration for with Knative Serving"
#./kn-operator remove resources -c serving -n ${SERVING_NAMESPACE} --deployName activator \
#  --container activator || fail_test "Failed to remove the resource configuration for Knative Serving"
#
#echo ">> Verify the resource configuration deletion for Knative Serving Custom resource"
#go_test_e2e -tags=servingresourceremove -timeout=20m ./test/e2e || failed=1
#
#echo ">> Remove the configMap configuration for with Knative Serving"
#./kn-operator remove configmaps -c serving -n ${SERVING_NAMESPACE} --cmName config-deployment \
#  --key ${TEST_KEY} || fail_test "Failed to remove the configmap configuration for Knative Serving"
#
#./kn-operator remove configmaps -c serving -n ${SERVING_NAMESPACE} --cmName config-network || fail_test "Failed to delete the ConfigMap configuration for Knative Serving"
#
#echo ">> Verify the configMap configuration deletion for Knative Serving Custom resource"
#go_test_e2e -tags=servingcmrremove -timeout=20m ./test/e2e || failed=1
#
#echo ">> Remove the toleration configuration for with Knative Serving"
#./kn-operator remove tolerations -c serving -n ${SERVING_NAMESPACE} --deployName autoscaler \
#  --key ${TOLERATION_KEY} || fail_test "Failed to remove the toleration configuration for Knative Serving"
#
#echo ">> Verify the toleration configuration deletion for Knative Serving Custom resource"
#go_test_e2e -tags=servingtolerationremove -timeout=20m ./test/e2e || failed=1
#
#echo ">> Delete the image of all deployments for Knative Seving"
#./kn-operator remove images -c serving -n ${SERVING_NAMESPACE} \
#  --imageKey default || fail_test "Failed to delete the image of all deployments for Knative Serving"
#
#echo ">> Delete the image of the deployment for Knative Seving"
#./kn-operator remove images -c serving -n ${SERVING_NAMESPACE} --deployName controller \
#  --imageKey ${SERVING_IMAGE_KEY} || fail_test "Failed to delete the image of the deployment for Knative Serving"
#
#echo ">> Verify the image deletion for Knative Seving"
#go_test_e2e -tags=servingimagedelete -timeout=20m ./test/e2e || failed=1

echo ">> Remove the number of replicas for Knative Serving"
./kn-operator remove replicas -c serving -n ${SERVING_NAMESPACE} --deployName controller || fail_test "Failed to remove the number of replias for Knative Serving"

./kn-operator remove replicas -c serving -n ${SERVING_NAMESPACE} || fail_test "Failed to remove the number of replias for Knative Serving"

echo ">> Verify the number of replicas for Knative Serving after removal"
go_test_e2e -tags=servingharemove -timeout=20m ./test/e2e || failed=1

#echo ">> Remove the environment variables for the container in the deployment of Knative Serving"
#./kn-operator remove envvars -c serving -n ${SERVING_NAMESPACE} --deployName controller --container controller \
#  --name ${ENV_NAME}|| fail_test "Failed to delete the env var for Knative Serving"
#
#./kn-operator remove envvars -c serving -n ${SERVING_NAMESPACE} --deployName controller --container controller \
#  --name ${ADDITIONAL_ENV_NAME}|| fail_test "Failed to delete the env var for Knative Serving"
#
#./kn-operator remove envvars -c serving -n ${SERVING_NAMESPACE} --deployName activator --container activator \
#  --name ${ENV_NAME} || fail_test "Failed to delete the env var for Knative Serving"
#
#echo ">> Verify the environment variables deletion for Knative Serving Custom resource"
#go_test_e2e -tags=servingenvvarsremove -timeout=20m ./test/e2e || failed=1
#
#echo ">> Remove the label for Knative Serving"
#./kn-operator remove labels -c serving -n ${SERVING_NAMESPACE} --deployName activator \
#  --key ${TEST_KEY} || fail_test "Failed to remove Knative Serving"
#
#./kn-operator remove labels -c serving -n ${SERVING_NAMESPACE} --deployName activator \
#  --key ${TEST_KEY_ADDITIONAL} || fail_test "Failed to remove Knative Serving"
#
#echo ">> Remove the annotation for Knative Serving"
#./kn-operator remove annotations -c serving -n ${SERVING_NAMESPACE} --deployName activator \
#  --key ${TEST_KEY} || fail_test "Failed to remove Knative Serving"
#
#./kn-operator remove annotations -c serving -n ${SERVING_NAMESPACE} --deployName activator \
#  --key ${TEST_KEY_ADDITIONAL} || fail_test "Failed to remove Knative Serving"
#
#echo ">> Remove the nodeSelector for Knative Serving"
#./kn-operator remove nodeSelectors -c serving -n ${SERVING_NAMESPACE} --deployName activator \
#  --key ${TEST_KEY} || fail_test "Failed to remove Knative Serving"
#
#./kn-operator remove nodeSelectors -c serving -n ${SERVING_NAMESPACE} --deployName activator \
#  --key ${TEST_KEY_ADDITIONAL} || fail_test "Failed to remove Knative Serving"
#
#echo ">> Verify the label, annotation and node selector deletion of Knative Serving"
#go_test_e2e -tags=servinglabeldelete -timeout=20m ./test/e2e || failed=1
#
#echo ">> Remove the label for Knative Serving's service"
#./kn-operator remove labels -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
#  --key ${TEST_KEY} || fail_test "Failed to remove Knative Serving's service"
#
#./kn-operator remove labels -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
#  --key ${TEST_KEY_ADDITIONAL} || fail_test "Failed to remove Knative Serving's service"
#
#echo ">> Remove the annotation for Knative Serving's service"
#./kn-operator remove annotations -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
#  --key ${TEST_KEY} || fail_test "Failed to remove Knative Serving's service"
#
#./kn-operator remove annotations -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
#  --key ${TEST_KEY_ADDITIONAL} || fail_test "Failed to remove Knative Serving's service"
#
#echo ">> Remove the selector for Knative Serving's service"
#./kn-operator remove selectors -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
#  --key ${TEST_KEY} || fail_test "Failed to remove Knative Serving's service"
#
#./kn-operator remove selectors -c serving -n ${SERVING_NAMESPACE} --serviceName activator-service \
#  --key ${TEST_KEY_ADDITIONAL} || fail_test "Failed to remove Knative Serving's service"
#
#echo ">> Verify the label, annotation selector deletion of Knative Serving's service"
#go_test_e2e -tags=servingservicelabeldelete -timeout=20m ./test/e2e || failed=1

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
  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure Knative Eventing"

./kn-operator configure labels -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure Knative Eventing"

echo ">> Configure the annotation for Knative Eventing"
./kn-operator configure annotations -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure Knative Eventing"

./kn-operator configure annotations -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure Knative Eventing"

echo ">> Configure the nodeSelector for Knative Eventing"
./kn-operator configure nodeSelectors -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure Knative Eventing"

./kn-operator configure nodeSelectors -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure Knative Eventing"

echo ">> Verify the label configuration of Knative Eventing"
go_test_e2e -tags=eventinglabelconfig -timeout=20m ./test/e2e || failed=1

echo ">> Configure the label for Knative Eventing's service"
./kn-operator configure labels -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure Knative Eventing's service"

./kn-operator configure labels -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure Knative Eventing's service"

echo ">> Configure the annotation for Knative Eventing's service"
./kn-operator configure annotations -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure Knative Eventing's service"

./kn-operator configure annotations -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure Knative Eventing's service"

echo ">> Configure the selector for Knative Eventing's service"
./kn-operator configure selectors -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY} --value ${TEST_VALUE} || fail_test "Failed to configure Knative Eventing's service"

./kn-operator configure selectors -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY_ADDITIONAL} --value ${TEST_VALUE_ADDITIONAL} || fail_test "Failed to configure Knative Eventing's service"

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

echo ">> Configure the environment variables for the container in the deployment of Knative Eventing"
./kn-operator configure envvars -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller --container eventing-controller \
  --name ${ENV_NAME} --value ${ENV_VALUE} || fail_test "Failed to configure the env var for Knative Eventing"

./kn-operator configure envvars -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller --container eventing-controller \
  --name ${ADDITIONAL_ENV_NAME} --value ${ADDITIONAL_ENV_VALUE} || fail_test "Failed to configure the env var for Knative Eventing"

./kn-operator configure envvars -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-webhook --container eventing-webhook \
  --name ${ENV_NAME} --value ${ENV_VALUE} || fail_test "Failed to configure the env var for Knative Eventing"

echo ">> Verify the env var configuration for Knative Eventing"
go_test_e2e -tags=eventingenvvar -timeout=20m ./test/e2e || failed=1

echo ">> Remove the resource configuration for with Knative Eventing"
./kn-operator remove resources -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --container eventing-controller || fail_test "Failed to remove the resource configuration for Knative Eventing"

echo ">> Verify the resource configuration deletion for Knative Eventing Custom resource"
go_test_e2e -tags=eventingresourcerremove -timeout=20m ./test/e2e || failed=1

echo ">> Remove the configMap configuration for with Knative Eventing"
./kn-operator remove configmaps -c eventing -n ${EVENTING_NAMESPACE} --cmName config-tracing \
  --key ${TEST_KEY} || fail_test "Failed to remove the configmap configuration for Knative Eventing"

./kn-operator remove configmaps -c eventing -n ${EVENTING_NAMESPACE} --cmName config-features || fail_test "Failed to delete the ConfigMap configuration for Knative Eventing"

echo ">> Verify the configMap configuration deletion for Knative Eventing Custom resource"
go_test_e2e -tags=eventingcmrremove -timeout=20m ./test/e2e || failed=1

echo ">> Delete the image of the deployment for Knative Eventing"
./kn-operator remove images -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --imageKey ${IMAGE_KEY} || fail_test "Failed to delete the image of the deployment for Knative Eventing"

echo ">> Delete the image of all deployments for Knative Eventing"
./kn-operator remove images -c eventing -n ${EVENTING_NAMESPACE} \
  --imageKey default || fail_test "Failed to delete the image of all deployments for Knative Eventing"

echo ">> Verify the image deletion for Knative Eventing"
go_test_e2e -tags=eventingimagedelete -timeout=20m ./test/e2e || failed=1

echo ">> Remove the number of replicas for Knative Eventing"
./kn-operator remove replicas -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller || fail_test "Failed to remove the number of replias for Knative Eventing"

./kn-operator remove replicas -c eventing -n ${EVENTING_NAMESPACE} || fail_test "Failed to remove the number of replias for Knative Eventing"

echo ">> Verify the number of replicas for Knative Eventing after removal"
go_test_e2e -tags=eventingharemove -timeout=20m ./test/e2e || failed=1

echo ">> Remove the environment variables for the container in the deployment of Knative Eventing"
./kn-operator remove envvars -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller --container eventing-controller \
  --name ${ENV_NAME} || fail_test "Failed to remove the env var for Knative Eventing"

./kn-operator remove envvars -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller --container eventing-controller \
  --name ${ADDITIONAL_ENV_NAME} || fail_test "Failed to remove the env var for Knative Eventing"

./kn-operator remove envvars -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-webhook --container eventing-webhook \
  --name ${ENV_NAME} || fail_test "Failed to delete the env var for Knative Eventing"

echo ">> Verify the environment variables deletion for Knative Eventing Custom resource"
go_test_e2e -tags=eventingenvvarsremove -timeout=20m ./test/e2e || failed=1

echo ">> Remove the label for Knative Eventing"
./kn-operator remove labels -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY} || fail_test "Failed to configure Knative Eventing"

./kn-operator remove labels -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY_ADDITIONAL} || fail_test "Failed to configure Knative Eventing"

echo ">> Remove the annotation for Knative Eventing"
./kn-operator remove annotations -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY} || fail_test "Failed to remove Knative Eventing"

./kn-operator remove annotations -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY_ADDITIONAL} || fail_test "Failed to remove Knative Eventing"

echo ">> Remove the nodeSelector for Knative Eventing"
./kn-operator remove nodeSelectors -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY} || fail_test "Failed to remove Knative Eventing"

./kn-operator remove nodeSelectors -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-controller \
  --key ${TEST_KEY_ADDITIONAL} || fail_test "Failed to remove Knative Eventing"

echo ">> Verify the label, annotation and nodeSelector deletion of Knative Eventing"
go_test_e2e -tags=eventinglabeldelete -timeout=20m ./test/e2e || failed=1

echo ">> Remove the label for Knative Eventing's service"
./kn-operator remove labels -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY} || fail_test "Failed to remove Knative Eventing's service"

./kn-operator remove labels -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY_ADDITIONAL} || fail_test "Failed to remove Knative Eventing's service"

echo ">> Remove the annotation for Knative Eventing's service"
./kn-operator remove annotations -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY} || fail_test "Failed to remove Knative Eventing's service"

./kn-operator remove annotations -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY_ADDITIONAL} || fail_test "Failed to remove Knative Eventing's service"

echo ">> Remove the selector for Knative Eventing's service"
./kn-operator remove selectors -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY} || fail_test "Failed to remove Knative Eventing's service"

./kn-operator remove selectors -c eventing -n ${EVENTING_NAMESPACE} --serviceName eventing-webhook \
  --key ${TEST_KEY_ADDITIONAL} || fail_test "Failed to remove Knative Eventing's service"

echo ">> Verify the label, annotation and selector deletion of Knative Eventing's service"
go_test_e2e -tags=eventingservicelabeldelete -timeout=20m ./test/e2e || failed=1

echo ">> Remove Knative Operator"
./kn-operator uninstall -n ${OPERATOR_NAMESPACE} || fail_test "Failed to remove Knative Operator"

echo ">> Remove the toleration configuration for with Knative Eventing"
./kn-operator remove tolerations -c eventing -n ${EVENTING_NAMESPACE} --deployName eventing-webhook \
  --key ${TOLERATION_KEY} || fail_test "Failed to remove the toleration configuration for Knative Eventing"

echo ">> Verify the toleration configuration deletion for Knative Eventing Custom resource"
go_test_e2e -tags=eventingtolerationremove -timeout=20m ./test/e2e || failed=1

(( failed )) && fail_test

success
