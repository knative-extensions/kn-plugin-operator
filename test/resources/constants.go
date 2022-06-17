package resources

import "os"

var (
	// ServingOperatorNamespace is the namespace for knative serving
	ServingOperatorNamespace = getenv("SERVING_NAMESPACE", "serving-test")
	// EventingOperatorNamespace is the namespace for knative eventing
	EventingOperatorNamespace = getenv("EVENTING_NAMESPACE", "eventing-test")
	// OperatorNamespace is the namespace of the Knative Operator
	OperatorNamespace = getenv("OPERATOR_NAMESPACE", "operator-test")
	// OperatorName is the name of the Knative Operator deployment
	OperatorName = "knative-operator"
	// TestKey is the key of the key-value pair for test cases
	TestKey = getenv("TEST_KEY", "test-key")
	// TestValue is the value of the key-value pair for test cases
	TestValue = getenv("TEST_VALUE", "test-value")
	// TestKeyAdditional is the additional key of the key-value pair for test cases
	TestKeyAdditional = getenv("TEST_KEY_ADDITIONAL", "test-key-additional")
	// TestValueAdditional is the additional value of the key-value pair for test cases
	TestValueAdditional = getenv("TEST_VALUE_ADDITIONAL", "test-value-additional")
	// TestReplicasNum is the number of replicas
	TestReplicasNum = getenv("REPLICA_NUM", "4")
	// TestTolerationKey is the test key for toleration
	TestTolerationKey = getenv("TOLERATION_KEY", "toleration-key")
	// TestOperation is the test operator
	TestOperation = getenv("OPERATION", "Exists")
	// TestEffect is the test effect
	TestEffect = getenv("EFFECT", "NoSchedule")
	// TestAdditionalTolerationKey is the additional test key for toleration
	TestAdditionalTolerationKey = getenv("ADDITIONAL_TOLERATION_KEY", "additional-toleration-key")
	// TestAdditionalOperation is the additional test operator
	TestAdditionalOperation = getenv("ADDITIONAL_OPERATION", "Equal")
	// TestAdditionalTolerationValue is the additional test value for toleration
	TestAdditionalTolerationValue = getenv("ADDITIONAL_TOLERATION_VALUE", "additional-toleration-value")
	// TestAdditionalEffect is the additional test effect for toleration
	TestAdditionalEffect = getenv("ADDITIONAL_EFFECT", "NoExecute")
	// TestImageUrl is the URL of the image
	TestImageUrl = getenv("IMAGE_URL", "gcr.io/knative-releases/knative.dev/eventing/cmd/controller:latest")
	// TestImageKey is the image key of the image
	TestImageKey = getenv("IMAGE_KEY", "eventing-controller")
	// TestServingImageUrl is the URL of the image
	TestServingImageUrl = getenv("SERVING_IMAGE_URL", "gcr.io/knative-releases/knative.dev/serving/cmd/controller:latest")
	// TestServingImageKey is the image key of the image
	TestServingImageKey = getenv("SERVING_IMAGE_KEY", "controller")
	// TestDefaultEventingImageUrl is the default eventing image url
	TestDefaultEventingImageUrl = getenv("DEFAULT_EVENTING_IMAGE_URL", "gcr.io/knative-releases/knative.dev/eventing/cmd/${NAME}:latest")
	// TestDefaultServingImageUrl is the default serving image url
	TestDefaultServingImageUrl = getenv("DEFAULT_SERVING_IMAGE_URL", "gcr.io/knative-releases/knative.dev/serving/cmd/${NAME}:latest")
	// TestEnvName is the name of the env var
	TestEnvName = getenv("ENV_NAME", "test-name")
	// TestEnvValue is the value of the env var
	TestEnvValue = getenv("ENV_VALUE", "test-value")
	// TestAddEnvName is the additional name of the env var
	TestAddEnvName = getenv("ADDITIONAL_ENV_NAME", "additional-test-name")
	// TestAddEnvValue is the additional value of the env var
	TestAddEnvValue = getenv("ADDITIONAL_ENV_VALUE", "additional-test-value")
)

func getenv(name, defaultValue string) string {
	value, set := os.LookupEnv(name)
	if !set {
		value = defaultValue
	}
	return value
}
