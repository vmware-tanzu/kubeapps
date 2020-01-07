module github.com/kubeapps/kubeapps

go 1.13

replace github.com/docker/docker => github.com/docker/docker v0.0.0-20190731150326-928381b2215c

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/arschles/assert v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-test/deep v1.0.4
	github.com/google/go-cmp v0.3.1
	github.com/gorilla/mux v1.7.3
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/kubeapps/common v0.0.0-20190508164739-10b110436c1a
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5
	github.com/unrolled/render v1.0.1 // indirect
	github.com/urfave/negroni v1.0.0
	google.golang.org/grpc v1.25.1
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	helm.sh/helm/v3 v3.0.0
	k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/helm v2.16.0+incompatible
	sigs.k8s.io/yaml v1.1.0
)
