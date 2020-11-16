module github.com/Project-Voltron/voltron/cmd/argo-runner

go 1.15

require (
	github.com/argoproj/argo v0.0.0-20201116043650-176d890c1cac
	github.com/davecgh/go-spew v1.1.1
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	github.com/vrischmann/envconfig v1.3.0
	go.uber.org/zap v1.16.0
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.5.11
	sigs.k8s.io/yaml v1.2.0
)

replace (
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.8
	k8s.io/client-go => k8s.io/client-go v0.17.8
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.11
)
