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
)

func getenv(name, defaultValue string) string {
	value, set := os.LookupEnv(name)
	if !set {
		value = defaultValue
	}
	return value
}
