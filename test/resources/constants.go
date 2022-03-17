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
)

func getenv(name, defaultValue string) string {
	value, set := os.LookupEnv(name)
	if !set {
		value = defaultValue
	}
	return value
}
