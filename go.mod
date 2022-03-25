module knative.dev/kn-plugin-operator

go 1.16

require (
	github.com/ghodss/yaml v1.0.0
	github.com/k14s/ytt v0.39.0
	github.com/manifestival/client-go-client v0.5.0
	github.com/spf13/cobra v1.3.0
	golang.org/x/mod v0.5.1
	k8s.io/api v0.23.4
	k8s.io/apimachinery v0.23.4
	k8s.io/client-go v0.23.4
	knative.dev/hack v0.0.0-20220318020218-14f832e506f8
	knative.dev/operator v0.30.1-0.20220323211118-b1e9c9d8ca0d
	knative.dev/pkg v0.0.0-20220318185521-e6e3cf03d765
)
