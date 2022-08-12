// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

module github.com/vmware-tanzu/kubeapps

go 1.18

replace (
	// required by https://github.com/kubernetes/code-generator/blob/master/go.mod
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.5

	// k8s.io/kubernetes is not intended to be used as a module, so versions are not being properly resolved.
	// This replacement is required, see https://github.com/kubernetes/kubernetes/issues/79384
	// As we support new k8s versions, this replacements should be also updated accordingly.
	k8s.io/api => k8s.io/api v0.22.10
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.10
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.10
	k8s.io/apiserver => k8s.io/apiserver v0.22.10
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.10
	k8s.io/client-go => k8s.io/client-go v0.22.10
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.22.10
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.22.10
	k8s.io/code-generator => k8s.io/code-generator v0.22.10
	k8s.io/component-base => k8s.io/component-base v0.22.10
	k8s.io/component-helpers => k8s.io/component-helpers v0.22.10
	k8s.io/controller-manager => k8s.io/controller-manager v0.22.10
	k8s.io/cri-api => k8s.io/cri-api v0.22.10
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.22.10
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.22.10
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.22.10
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.22.10
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.22.10
	k8s.io/kubectl => k8s.io/kubectl v0.22.10
	k8s.io/kubelet => k8s.io/kubelet v0.22.10
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.22.10
	k8s.io/metrics => k8s.io/metrics v0.22.10
	k8s.io/mount-utils => k8s.io/mount-utils v0.22.10
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.22.10
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.22.10
)

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/ahmetb/go-linq/v3 v3.2.0
	github.com/containerd/containerd v1.6.6
	github.com/cppforlife/go-cli-ui v0.0.0-20220622150351-995494831c6c
	github.com/disintegration/imaging v1.6.2
	github.com/distribution/distribution v2.8.1+incompatible
	github.com/fluxcd/helm-controller/api v0.22.2
	github.com/fluxcd/pkg/apis/meta v0.14.2
	github.com/fluxcd/source-controller/api v0.26.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-redis/redismock/v8 v8.0.6
	github.com/google/go-cmp v0.5.8
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.0
	github.com/heptiolabs/healthcheck v0.0.0-20211123025425-613501dd5deb
	github.com/improbable-eng/grpc-web v0.15.0
	github.com/itchyny/gojq v0.12.8
	github.com/jinzhu/copier v0.3.5
	github.com/k14s/kapp v0.50.0
	github.com/lib/pq v1.10.6
	github.com/mitchellh/go-homedir v1.1.0
	github.com/opencontainers/image-spec v1.0.3-0.20211202183452-c5a74bcca799
	github.com/soheilhy/cmux v0.1.5
	github.com/spf13/cobra v1.5.0
	github.com/spf13/cobra-cli v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.12.0
	github.com/srwiley/oksvg v0.0.0-20220128195007-1f435e4c2b44
	github.com/srwiley/rasterx v0.0.0-20220615024203-67b7089efd25
	github.com/stretchr/testify v1.8.0
	github.com/urfave/negroni/v2 v2.0.2
	github.com/vmware-tanzu/carvel-kapp-controller v0.38.4
	github.com/vmware-tanzu/carvel-vendir v0.29.0
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4
	google.golang.org/genproto v0.0.0-20220725144611-272f38e5d71b
	google.golang.org/grpc v1.48.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.2.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.8.2
	k8s.io/api v0.24.0
	k8s.io/apiextensions-apiserver v0.24.1
	k8s.io/apimachinery v0.24.1
	k8s.io/apiserver v0.23.5
	k8s.io/cli-runtime v0.23.5
	k8s.io/client-go v0.23.5
	k8s.io/klog/v2 v2.70.1
	k8s.io/kubectl v0.23.5
	k8s.io/kubernetes v1.22.10
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
	oras.land/oras-go v1.2.0
	oras.land/oras-go/v2 v2.0.0-rc.1
	sigs.k8s.io/controller-runtime v0.11.2
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/BurntSushi/toml v1.2.0 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.2 // indirect
	github.com/Masterminds/squirrel v1.5.3 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/chai2010/gettext-go v0.1.0 // indirect
	github.com/cppforlife/cobrautil v0.0.0-20220411122935-c28a9f274a4e // indirect
	github.com/cppforlife/color v1.9.1-0.20200716202919-6706ac40b835 // indirect
	github.com/cppforlife/go-patch v0.2.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/docker/cli v20.10.17+incompatible
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.17+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.4 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fluxcd/pkg/apis/acl v0.0.3 // indirect
	github.com/fluxcd/pkg/apis/kustomize v0.4.2 // indirect
	github.com/fluxcd/pkg/version v0.1.0
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/fvbommel/sortorder v1.0.2 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/gin-gonic/gin v1.8.1 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-gorp/gorp/v3 v3.0.2 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gnostic v0.6.9 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/itchyny/timefmt-go v0.1.3 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/k14s/difflib v0.0.0-20201117154628-0c031775bf57 // indirect
	github.com/k14s/semver/v4 v4.0.1-0.20210701191048-266d47ac6115 // indirect
	github.com/k14s/starlark-go v0.0.0-20200720175618-3a5c849cc368 // indirect
	github.com/k14s/ytt v0.39.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/distribution-spec/specs-go v0.0.0-20220620172159-4ab4752c3b86 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/oras-project/artifacts-spec v1.0.0-rc.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.2 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.2 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rs/cors v1.8.2 // indirect
	github.com/rubenv/sql-migrate v1.1.2 // indirect
	github.com/russross/blackfriday v1.6.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/spf13/afero v1.9.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.4.0 // indirect
	github.com/vito/go-interact v1.0.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.1.0 // indirect
	go.starlark.net v0.0.0-20220714194419-4cadf0a12139 // indirect
	golang.org/x/crypto v0.0.0-20220722155217-630584e8d5aa // indirect
	golang.org/x/image v0.0.0-20220722155232-062f8c9fd539 // indirect
	golang.org/x/oauth2 v0.0.0-20220722155238-128564f6959c // indirect
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f // indirect
	golang.org/x/term v0.0.0-20220722155259-a9ba230a4035 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220722155302-e5dcc9cfc0b9 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.6 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/component-base v0.23.5 // indirect
	k8s.io/kube-openapi v0.0.0-20220124234850-424119656bbf // indirect
	nhooyr.io/websocket v1.8.7 // indirect
	sigs.k8s.io/kustomize/api v0.11.1 // indirect
	sigs.k8s.io/kustomize/kyaml v0.13.3 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
)
