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
)

func getenv(name, defaultValue string) string {
	value, set := os.LookupEnv(name)
	if !set {
		value = defaultValue
	}
	return value
}
