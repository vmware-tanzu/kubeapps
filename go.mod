module github.com/kubeapps/kubeapps

go 1.16

replace (
	// required by Oras https://github.com/deislabs/oras/blob/main/go.mod
	github.com/docker/distribution => github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker => github.com/moby/moby v20.10.6+incompatible

	// k8s.io/kubernetes is not intended to be used as a module, so versions are not being properly resolved.
	// This replacement is required, see https://github.com/kubernetes/kubernetes/issues/79384
	// As we support new k8s versions, this replacements should be also updated accordingly.
	k8s.io/api => k8s.io/api v0.20.5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.5
	k8s.io/apiserver => k8s.io/apiserver v0.20.5
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.5
	k8s.io/client-go => k8s.io/client-go v0.20.5
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.5
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.5
	k8s.io/code-generator => k8s.io/code-generator v0.20.5
	k8s.io/component-base => k8s.io/component-base v0.20.5
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.5
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.5
	k8s.io/cri-api => k8s.io/cri-api v0.20.5
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.5
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.5
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.5
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.5
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.5
	k8s.io/kubectl => k8s.io/kubectl v0.20.5
	k8s.io/kubelet => k8s.io/kubelet v0.20.5
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.5
	k8s.io/metrics => k8s.io/metrics v0.20.5
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.5
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.5
)

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d // indirect
	github.com/arschles/assert v1.0.0
	github.com/bshuster-repo/logrus-logstash-hook v1.0.0 // indirect
	github.com/bugsnag/bugsnag-go v1.5.0 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/containerd/containerd v1.4.4
	github.com/deislabs/oras v0.11.1
	github.com/disintegration/imaging v1.6.2
	github.com/distribution/distribution v2.7.1+incompatible
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/spec v0.19.4 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6
	github.com/gorilla/handlers v1.4.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.5.0
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/itchyny/gojq v0.12.4
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/kardianos/osext v0.0.0-20170510131534-ae77be60afb1 // indirect
	github.com/kubeapps/common v0.0.0-20200304064434-f6ba82e79f47
	github.com/lib/pq v1.10.2
	github.com/mitchellh/go-homedir v1.1.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/soheilhy/cmux v0.1.5
	github.com/spf13/cobra v1.2.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/srwiley/oksvg v0.0.0-20210320200257-875f767ac39a
	github.com/srwiley/rasterx v0.0.0-20200120212402-85cb7272f5e9
	github.com/stretchr/testify v1.7.0
	github.com/unrolled/render v1.0.1 // indirect
	github.com/urfave/negroni v1.0.0
	github.com/yvasiyarov/go-metrics v0.0.0-20150112132944-c25f46c4b940 // indirect
	github.com/yvasiyarov/gorelic v0.0.6 // indirect
	github.com/yvasiyarov/newrelic_platform_go v0.0.0-20160601141957-9c099fbc30e9 // indirect
	golang.org/x/image v0.0.0-20201208152932-35266b937fa6 // indirect
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	google.golang.org/genproto v0.0.0-20210617175327-b9e0b3197ced
	google.golang.org/grpc v1.39.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.5.4
	k8s.io/api v0.20.5
	k8s.io/apimachinery v0.20.5
	k8s.io/cli-runtime v0.20.5
	k8s.io/client-go v0.20.5
	k8s.io/helm v2.17.0+incompatible
	k8s.io/klog/v2 v2.9.0
	k8s.io/kubernetes v1.20.5
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/yaml v1.2.0
)
