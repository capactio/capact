module projectvoltron.dev/voltron/docs/investigation/workflow-rendering

go 1.15

require (
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible
	github.com/argoproj/argo v2.5.2+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/mitchellh/mapstructure v1.4.0
	github.com/pkg/errors v0.9.1
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.20.1 // indirect
	projectvoltron.dev/voltron v0.1.0
	sigs.k8s.io/yaml v1.2.0
)

replace projectvoltron.dev/voltron => ../../..

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.11
