module projectvoltron.dev/voltron

go 1.15

require (
	github.com/99designs/gqlgen v0.13.0
	github.com/99designs/keyring v1.1.6
	github.com/AlecAivazis/survey/v2 v2.2.8
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd
	github.com/Masterminds/semver/v3 v3.0.3
	github.com/agnivade/levenshtein v1.1.0 // indirect
	github.com/argoproj/argo/v2 v2.12.10
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/common-nighthawk/go-figure v0.0.0-20200609044655-c4b36f998cf2
	github.com/docker/cli v0.0.0-20200130152716-5d0cf8839492
	github.com/docker/docker v1.4.2-0.20200203170920-46ec8731fbce
	github.com/fatih/color v1.10.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.0
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-getter v1.5.2
	github.com/hashicorp/hcl/v2 v2.9.1
	github.com/hashicorp/terraform v0.14.8
	github.com/hokaccha/go-prettyjson v0.0.0-20210113012101-fb4e108d2519
	github.com/iancoleman/strcase v0.1.2
	github.com/machinebox/graphql v0.2.2
	github.com/matryer/is v1.4.0 // indirect
	github.com/mattn/go-runewidth v0.0.10 // indirect
	github.com/nautilus/gateway v0.1.7
	github.com/nautilus/graphql v0.0.12
	github.com/neo4j/neo4j-go-driver/v4 v4.2.2
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sethvargo/go-password v0.2.0
	github.com/shurcooL/httpfs v0.0.0-20171119174359-809beceb2371
	github.com/shurcooL/vfsgen v0.0.0-20180121065927-ffb13db8def0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
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
	k8s.io/apimachinery v0.18.2
	k8s.io/cli-runtime v0.17.9
	k8s.io/client-go v10.0.0+incompatible
	k8s.io/kubectl v0.17.2
	k8s.io/utils v0.0.0-20200619165400-6e3d28b6ed19
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.0.0-00010101000000-000000000000
	sigs.k8s.io/yaml v1.2.0
)

replace (
	// Remove when the is resolved:
	// - https://github.com/argoproj/argo-workflows/issues/4772
	// - we can compile argo without static files
	github.com/argoproj/argo/v2 => github.com/Project-Voltron/argo-workflows/v2 v2.12.10-0.20210323093745-be9145c858b1
	// Remove when the issues are resolved:
	// https://github.com/graphql-go/graphql/issues/586
	github.com/graphql-go/graphql => github.com/pkosiec/graphql-go v0.7.10-0.20201208110622-388f8a2d4f19

	github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
	// https://github.com/nautilus/gateway/issues/121
	github.com/nautilus/graphql => github.com/pkosiec/graphql v0.0.13-0.20201208111257-86f2e16b2778

	// Remove after vendoring new Argo: https://github.com/argoproj/argo/pull/4426
	k8s.io/api => k8s.io/api v0.17.9
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.9
	k8s.io/client-go => k8s.io/client-go v0.17.9
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.11
)
