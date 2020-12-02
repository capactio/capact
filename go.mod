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
	github.com/google/go-cmp v0.5.1 // indirect
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.7.2
	github.com/iancoleman/strcase v0.1.2
	github.com/machinebox/graphql v0.2.2
	github.com/matryer/is v1.4.0 // indirect
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/nautilus/gateway v0.1.7
	github.com/nautilus/graphql v0.0.12
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pkg/errors v0.9.1
<<<<<<< HEAD
	github.com/shurcooL/httpfs v0.0.0-20171119174359-809beceb2371
	github.com/shurcooL/vfsgen v0.0.0-20180121065927-ffb13db8def0
=======
	github.com/sethvargo/go-password v0.2.0
>>>>>>> d784a0d... json schemas
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/vektah/gqlparser/v2 v2.1.0
	github.com/vrischmann/envconfig v1.3.0
	github.com/xeipuuv/gojsonschema v1.2.0
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/api v0.20.0
	gotest.tools v2.2.0+incompatible
	helm.sh/helm/v3 v3.1.2
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/cli-runtime v0.17.9
	k8s.io/client-go v0.17.9
	rsc.io/letsencrypt v0.0.3 // indirect
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
