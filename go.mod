module capact.io/capact

go 1.16

require (
	github.com/99designs/gqlgen v0.13.0
	github.com/99designs/keyring v1.1.6
	github.com/AlecAivazis/survey/v2 v2.2.16
	github.com/BurntSushi/toml v0.4.1 // indirect
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd
	github.com/Masterminds/goutils v1.1.1
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/alecthomas/jsonschema v0.0.0-20210526225647-edb03dcab7bc
	github.com/argoproj/argo-workflows/v3 v3.1.0-rc1.0.20210811221840-88520891a037
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/aws/aws-sdk-go v1.37.0 // indirect
	github.com/briandowns/spinner v1.12.0
	github.com/common-nighthawk/go-figure v0.0.0-20200609044655-c4b36f998cf2
	github.com/docker/cli v20.10.6+incompatible
	github.com/docker/docker v20.10.8+incompatible
	github.com/evanphx/json-patch/v5 v5.5.0 // indirect
	github.com/fatih/color v1.12.0
	github.com/fatih/structs v1.1.0
	github.com/gitchander/permutation v0.0.0-20210302120832-6ab79d7de174
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-getter v1.5.5
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/hcl/v2 v2.9.1
	github.com/hashicorp/hcl2 v0.0.0-20191002203319-fb75b3253c80 // indirect
	github.com/hashicorp/terraform v0.11.12-beta1
	github.com/hashicorp/terraform-config-inspect v0.0.0-20210625153042-09f34846faab
	github.com/hokaccha/go-prettyjson v0.0.0-20210113012101-fb4e108d2519
	github.com/iancoleman/orderedmap v0.0.0-20190318233801-ac98e3ecb4b0
	github.com/iancoleman/strcase v0.1.2
	github.com/machinebox/graphql v0.2.2
	github.com/matryer/is v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.13
	github.com/mitchellh/mapstructure v1.4.1
	github.com/nautilus/gateway v0.1.16
	github.com/nautilus/graphql v0.0.16
	github.com/neo4j/neo4j-go-driver/v4 v4.2.2
	github.com/olekukonko/tablewriter v0.0.4
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	github.com/pkg/errors v0.9.1
	github.com/rancher/k3d/v4 v4.4.8
	github.com/sethvargo/go-password v0.2.0
	github.com/shurcooL/httpfs v0.0.0-20171119174359-809beceb2371
	github.com/shurcooL/vfsgen v0.0.0-20180121065927-ffb13db8def0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/theupdateframework/notary v0.7.0 // indirect
	github.com/valyala/fastjson v1.6.3
	github.com/vektah/gqlparser/v2 v2.1.0
	github.com/vrischmann/envconfig v1.3.0
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/zclconf/go-cty v1.8.1
	go.uber.org/zap v1.18.1
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e // indirect
	google.golang.org/api v0.44.0
	google.golang.org/grpc/examples v0.0.0-20210322221411-d26af8e39165 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gotest.tools v2.2.0+incompatible
	helm.sh/helm/v3 v3.6.3
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/cli-runtime v0.21.0
	k8s.io/client-go v0.21.3
	k8s.io/kubectl v0.21.0
	k8s.io/utils v0.0.0-20210802155522-efc7438f0176
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.9.6
	sigs.k8s.io/kind v0.11.1
	sigs.k8s.io/yaml v1.2.0
)

replace (
	// TODO:
	// 	- Remove when we can compile argo without static files
	// 	- Use stable tag once new version with Kubernetes 1.21 usage is released
	github.com/argoproj/argo-workflows/v3 => github.com/capactio/argo-workflows/v3 v3.1.0-rc1.0.20210812143110-6bc7f066ec1f

	github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.8
	github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
)
