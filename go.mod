module github.com/kubeapps/kubeapps

go 1.13

replace github.com/docker/docker => github.com/docker/docker v0.0.0-20190731150326-928381b2215c

require (
	github.com/DATA-DOG/go-sqlmock v1.3.3
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/arschles/assert v1.0.0
	github.com/disintegration/imaging v1.6.2
	github.com/ghodss/yaml v1.0.0
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/go-test/deep v1.0.4
	github.com/golang/protobuf v1.3.2
	github.com/google/go-cmp v0.3.1
	github.com/gorilla/mux v1.7.3
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/kubeapps/common v0.0.0-20200304064434-f6ba82e79f47
	github.com/lib/pq v1.2.0
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/unrolled/render v1.0.1 // indirect
	github.com/urfave/negroni v1.0.0
	google.golang.org/grpc v1.25.1
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/yaml.v2 v2.2.4
	helm.sh/helm/v3 v3.0.2
	k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/helm v2.16.0+incompatible
	sigs.k8s.io/yaml v1.1.0
)
