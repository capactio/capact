module capact.io/capact

go 1.16

require (
	github.com/99designs/gqlgen v0.13.0
	github.com/99designs/keyring v1.1.6
	github.com/AlecAivazis/survey/v2 v2.2.9
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd
	github.com/Masterminds/semver/v3 v3.0.3
	github.com/argoproj/argo/v2 v2.0.0-00010101000000-000000000000
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/briandowns/spinner v1.12.0
	github.com/common-nighthawk/go-figure v0.0.0-20200609044655-c4b36f998cf2
	github.com/docker/cli v0.0.0-20200130152716-5d0cf8839492
	github.com/docker/docker v1.4.2-0.20200203170920-46ec8731fbce
	github.com/fatih/color v1.10.0
	github.com/fatih/structs v1.1.0
	github.com/gitchander/permutation v0.0.0-20210302120832-6ab79d7de174
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.0
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-getter v1.5.5
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/hcl/v2 v2.9.1
	github.com/hashicorp/terraform v0.14.8
	github.com/hokaccha/go-prettyjson v0.0.0-20210113012101-fb4e108d2519
	github.com/iancoleman/strcase v0.1.2
	github.com/machinebox/graphql v0.2.2
	github.com/matryer/is v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.12
	github.com/mitchellh/mapstructure v1.4.1
	github.com/nautilus/gateway v0.1.16
	github.com/nautilus/graphql v0.0.16
	github.com/neo4j/neo4j-go-driver/v4 v4.2.2
	github.com/olekukonko/tablewriter v0.0.0-20170122224234-a0225b3f23b5
	github.com/onsi/ginkgo v1.15.1
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/sethvargo/go-password v0.2.0
	github.com/shurcooL/httpfs v0.0.0-20171119174359-809beceb2371
	github.com/shurcooL/vfsgen v0.0.0-20180121065927-ffb13db8def0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/theupdateframework/notary v0.7.0 // indirect
	github.com/vektah/gqlparser/v2 v2.1.0
	github.com/vrischmann/envconfig v1.3.0
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/zclconf/go-cty v1.8.1
	go.uber.org/zap v1.10.0
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/sys v0.0.0-20210309074719-68d13333faf2 // indirect
	google.golang.org/api v0.34.0
	google.golang.org/grpc/examples v0.0.0-20210322221411-d26af8e39165 // indirect
	gotest.tools v2.2.0+incompatible
	helm.sh/helm/v3 v3.1.2
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.20.2
	k8s.io/cli-runtime v0.17.9
	k8s.io/client-go v10.0.0+incompatible
	k8s.io/helm v2.17.0+incompatible
	k8s.io/kubectl v0.17.2
	k8s.io/utils v0.0.0-20200619165400-6e3d28b6ed19
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.0.0-00010101000000-000000000000
	sigs.k8s.io/kind v0.11.1
	sigs.k8s.io/yaml v1.2.0
)

replace (
	// Remove when the is resolved:
	// - https://github.com/argoproj/argo-workflows/issues/4772
	// - we can compile argo without static files
	github.com/argoproj/argo/v2 => github.com/capactio/argo-workflows/v2 v2.12.10-0.20210323093745-be9145c858b1

	github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4

	// Remove after vendoring new Argo: https://github.com/argoproj/argo/pull/4426
	k8s.io/api => k8s.io/api v0.17.9
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.9
	k8s.io/client-go => k8s.io/client-go v0.17.9
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.11
)
