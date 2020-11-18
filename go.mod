module projectvoltron.dev/voltron

go 1.15

require (
	github.com/99designs/gqlgen v0.13.0
	github.com/argoproj/argo v0.0.0-20201118180151-53195ed56029
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/fatih/color v1.10.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.0
	github.com/gorilla/mux v1.6.1
	github.com/iancoleman/strcase v0.1.2
	github.com/nautilus/gateway v0.1.9
	github.com/nautilus/graphql v0.0.12
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.6.1
	github.com/vektah/gqlparser/v2 v2.1.0
	github.com/vrischmann/envconfig v1.3.0
	github.com/xeipuuv/gojsonschema v1.2.0
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.10.0
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.17.9
	sigs.k8s.io/controller-runtime v0.0.0-00010101000000-000000000000
	sigs.k8s.io/yaml v1.2.0
)

// Can be removed after vendoring new Argo: https://github.com/argoproj/argo/pull/4426
replace (
	k8s.io/api => k8s.io/api v0.17.9
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.9
	k8s.io/client-go => k8s.io/client-go v0.17.9
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.11
)
