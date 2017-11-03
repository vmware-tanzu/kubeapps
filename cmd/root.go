package cmd

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ksonnet/kubecfg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	KUBEAPPSNS = `---
apiVersion: v1
kind: Namespace
metadata:
  annotations: {}
  labels:
    name: kubeapps
  name: kubeapps
`

	MANIFEST = `---
apiVersion: v1
data:
  mongodb-password: MjNneWZ3ZWZoZzkyOA==
  mongodb-root-password: MjNneWZ3ZWZoZzkyOA==
kind: Secret
metadata:
  annotations: {}
  name: mongodb
  namespace: kubeapps
type: Opaque
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: helm
    name: tiller
  name: tiller-deploy
  namespace: kube-system
spec:
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: helm
        name: tiller
    spec:
      containers:
      - env:
        - name: TILLER_NAMESPACE
          value: kube-system
        - name: TILLER_HISTORY_MAX
          value: "0"
        image: gcr.io/kubernetes-helm/tiller:v2.7.0
        imagePullPolicy: IfNotPresent
        livenessProbe:
          httpGet:
            path: /liveness
            port: 44135
          initialDelaySeconds: 1
          timeoutSeconds: 1
        name: tiller
        ports:
        - containerPort: 44134
          name: tiller
        readinessProbe:
          httpGet:
            path: /readiness
            port: 44135
          initialDelaySeconds: 1
          timeoutSeconds: 1
        resources: {}
status: {}
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  annotations: {}
  labels:
    app: kubeapps-hub
    name: kubeapps-hub-prerender
  name: kubeapps-hub-prerender
  namespace: kubeapps
spec:
  minReadySeconds: 30
  replicas: 1
  selector:
    matchLabels:
      app: kubeapps-hub
      name: kubeapps-hub-prerender
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      annotations: {}
      labels:
        app: kubeapps-hub
        name: kubeapps-hub-prerender
    spec:
      containers:
      - args: []
        env:
        - name: IN_MEMORY_CACHE
          value: "true"
        image: migmartri/prerender:latest
        imagePullPolicy: Always
        name: prerender
        ports:
        - containerPort: 3000
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
        stdin: false
        tty: false
        volumeMounts: []
      imagePullSecrets: []
      initContainers: []
      terminationGracePeriodSeconds: 30
      volumes: []
---
apiVersion: v1
kind: Service
metadata:
  annotations: {}
  labels:
    app: kubeapps-hub
    name: kubeapps-hub-prerender
  name: kubeapps-hub-prerender
  namespace: kubeapps
spec:
  ports:
  - name: prerender
    port: 80
    protocol: TCP
    targetPort: 3000
  selector:
    app: kubeapps-hub
    name: kubeapps-hub-prerender
  type: NodePort
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  annotations: {}
  labels:
    app: kubeapps-hub
    name: kubeapps-hub-ratesvc
  name: kubeapps-hub-ratesvc
  namespace: kubeapps
spec:
  minReadySeconds: 30
  replicas: 1
  selector:
    matchLabels:
      app: kubeapps-hub
      name: kubeapps-hub-ratesvc
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      annotations: {}
      labels:
        app: kubeapps-hub
        name: kubeapps-hub-ratesvc
    spec:
      containers:
      - args:
        - /ratesvc
        - -mongo-host
        - mongodb.kubeapps
        - -mongo-database
        - ratesvc
        env:
        - name: JWT_KEY
          value: secret
        image: kubeapps/ratesvc:v0.1.0
        imagePullPolicy: Always
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
        name: ratesvc
        ports:
        - containerPort: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
        resources: {}
        stdin: false
        tty: false
        volumeMounts: []
      imagePullSecrets: []
      initContainers: []
      terminationGracePeriodSeconds: 30
      volumes: []
---
apiVersion: v1
kind: Service
metadata:
  annotations: {}
  labels:
    app: kubeapps-hub
    name: kubeapps-hub-ratesvc
  name: kubeapps-hub-ratesvc
  namespace: kubeapps
spec:
  ports:
  - name: ratesvc
    port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app: kubeapps-hub
    name: kubeapps-hub-ratesvc
  type: ClusterIP
---
apiVersion: v1
data:
  overrides.js: |
    window.monocular = {
        "overrides": {
            "appName": "Monocular",
            "backendHostname": "/api",
            "googleAnalyticsId": "UA-XXXXXX-X",
            "releasesEnabled": true
        }
    }
kind: ConfigMap
metadata:
  annotations: {}
  labels:
    app: kubeapps-hub
    name: kubeapps-hub-ui-config
  name: kubeapps-hub-ui-config-024dc17
  namespace: kubeapps
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  annotations: {}
  labels:
    app: kubeapps-hub
    name: kubeapps-hub-ui
  name: kubeapps-hub-ui
  namespace: kubeapps
spec:
  minReadySeconds: 30
  replicas: 2
  selector:
    matchLabels:
      app: kubeapps-hub
      name: kubeapps-hub-ui
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      annotations: {}
      labels:
        app: kubeapps-hub
        name: kubeapps-hub-ui
    spec:
      containers:
      - args: []
        env: []
        image: kubeapps/hub:latest
        imagePullPolicy: Always
        livenessProbe:
          httpGet:
            path: /
            port: 8080
          initialDelaySeconds: 60
          timeoutSeconds: 10
        name: ui
        ports:
        - containerPort: 8080
        readinessProbe:
          httpGet:
            path: /
            port: 8080
          initialDelaySeconds: 30
          timeoutSeconds: 5
        resources:
          limits:
            cpu: 100m
            memory: 128Mi
          requests:
            cpu: 100m
            memory: 128Mi
        stdin: false
        tty: false
        volumeMounts:
        - mountPath: /app/assets/js
          name: config
        - mountPath: /bitnami/nginx/conf/vhosts
          name: vhost
      imagePullSecrets: []
      initContainers: []
      terminationGracePeriodSeconds: 30
      volumes:
      - configMap:
          name: kubeapps-hub-ui-config-024dc17
        name: config
      - configMap:
          name: kubeapps-hub-ui-vhost-a3dec15
        name: vhost
---
apiVersion: v1
kind: Service
metadata:
  annotations: {}
  labels:
    app: kubeapps-hub
    name: kubeapps-hub-ui
  name: kubeapps-hub-ui
  namespace: kubeapps
spec:
  ports:
  - name: monocular-ui
    port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app: kubeapps-hub
    name: kubeapps-hub-ui
  type: NodePort
---
apiVersion: v1
data:
  vhost.conf: |
    upstream target_service {
      server kubeapps-hub-prerender;
    }

    server {
      listen 8080;

      gzip on;
      # Angular CLI already has gzipped the assets (ng build --prod --aot)
      gzip_static  on;

      location / {
        try_files $uri @prerender;
      }
      location @prerender {
        set $prerender 0;
        if ($http_user_agent ~* "baiduspider|twitterbot|facebookexternalhit|rogerbot|linkedinbot|embedly|quora link preview|showyoubot|outbrain|pinterest|slackbot|vkShare|W3C_Validator") {
          set $prerender 1;
        }
        if ($args ~ "_escaped_fragment_") {
          set $prerender 1;
        }
        if ($http_user_agent ~ "Prerender") {
          set $prerender 0;
        }
        if ($uri ~* "\.(js|css|xml|less|png|jpg|jpeg|gif|pdf|doc|txt|ico|rss|zip|mp3|rar|exe|wmv|doc|avi|ppt|mpg|mpeg|tif|wav|mov|psd|ai|xls|mp4|m4a|swf|dat|dmg|iso|flv|m4v|torrent|ttf|woff|svg|eot)") {
          set $prerender 0;
        }
        if ($prerender = 1) {
          rewrite .* /https://$host$request_uri? break;
          proxy_pass http://target_service;
        }
        if ($prerender = 0) {
          rewrite .* /index.html break;
        }
      }
    }

    # Redirect www to non-www
    # Taken from https://easyengine.io/tutorials/nginx/www-non-www-redirection/
    server {
      server_name "~^www\.(.*)$" ;
      return 301 $scheme://$1$request_uri ;
    }
kind: ConfigMap
metadata:
  annotations: {}
  labels:
    app: kubeapps-hub
    name: kubeapps-hub-ui-vhost
  name: kubeapps-hub-ui-vhost-a3dec15
  namespace: kubeapps
---
apiVersion: v1
data:
  monocular.yaml: |
    {
        "cacheRefreshInterval": 3600,
        "cors": {
            "allowed_headers": [
                "content-type",
                "x-xsrf-token"
            ],
            "allowed_origins": [
                ""
            ]
        },
        "mongodb": {
            "database": "monocular",
            "host": "mongodb.kubeapps:27017"
        },
        "releasesEnabled": true,
        "repos": [
            {
                "name": "stable",
                "source": "https://github.com/kubernetes/charts/tree/master/stable",
                "url": "https://kubernetes-charts.storage.googleapis.com"
            },
            {
                "name": "incubator",
                "source": "https://github.com/kubernetes/charts/tree/master/incubator",
                "url": "https://kubernetes-charts-incubator.storage.googleapis.com"
            }
        ]
    }
kind: ConfigMap
metadata:
  annotations: {}
  labels:
    app: kubeapps-hub
    name: kubeapps-hub-api
  name: kubeapps-hub-api-4f1c258
  namespace: kubeapps
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  annotations: {}
  labels:
    app: kubeapps-hub
    name: kubeapps-hub-api
  name: kubeapps-hub-api
  namespace: kubeapps
spec:
  minReadySeconds: 30
  replicas: 2
  selector:
    matchLabels:
      app: kubeapps-hub
      name: kubeapps-hub-api
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      annotations: {}
      labels:
        app: kubeapps-hub
        name: kubeapps-hub-api
    spec:
      containers:
      - args: []
        env:
        - name: MONOCULAR_AUTH_GITHUB_CLIENT_ID
          value: "null"
        - name: MONOCULAR_AUTH_GITHUB_CLIENT_SECRET
          value: "null"
        - name: MONOCULAR_AUTH_SIGNING_KEY
          value: secret
        - name: MONOCULAR_HOME
          value: /monocular
        image: bitnami/monocular-api:v0.5.2
        imagePullPolicy: Always
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8081
          initialDelaySeconds: 180
          timeoutSeconds: 10
        name: api
        ports:
        - containerPort: 8081
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8081
          initialDelaySeconds: 30
          timeoutSeconds: 5
        resources:
          limits:
            cpu: 100m
            memory: 128Mi
          requests:
            cpu: 100m
            memory: 128Mi
        stdin: false
        tty: false
        volumeMounts:
        - mountPath: /monocular
          name: cache
        - mountPath: /monocular/config
          name: config
      imagePullSecrets: []
      initContainers: []
      terminationGracePeriodSeconds: 30
      volumes:
      - emptyDir: {}
        name: cache
      - configMap:
          name: kubeapps-hub-api-4f1c258
        name: config
---
apiVersion: v1
kind: Service
metadata:
  annotations: {}
  labels:
    app: kubeapps-hub
    name: kubeapps-hub-api
  name: kubeapps-hub-api
  namespace: kubeapps
spec:
  ports:
  - name: monocular-api
    port: 80
    protocol: TCP
    targetPort: 8081
  selector:
    app: kubeapps-hub
    name: kubeapps-hub-api
  type: NodePort
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    ingress.kubernetes.io/rewrite-target: /
    kubernetes.io/ingress.class: nginx
  labels:
    app: kubeapps-hub
    name: kubeapps-hub
  name: kubeapps-hub
  namespace: kubeapps
spec:
  rules:
  - host: null
    http:
      paths:
      - backend:
          serviceName: kubeapps-hub-ui
          servicePort: 80
        path: /
      - backend:
          serviceName: kubeapps-hub-api
          servicePort: 80
        path: /api/
      - backend:
          serviceName: kubeapps-hub-ratesvc
          servicePort: 80
        path: /api/ratesvc
  tls: []
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: kubeless-controller-deployer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kubeless-controller-deployer
subjects:
- kind: ServiceAccount
  name: controller-acct
  namespace: kubeless
---
apiVersion: v1
kind: Service
metadata:
  name: kafka
  namespace: kubeless
spec:
  ports:
  - port: 9092
  selector:
    kubeless: kafka
---
apiVersion: v1
kind: Namespace
metadata:
  annotations: {}
  labels:
    name: kubeless
  name: kubeless
---
apiVersion: v1
kind: Service
metadata:
  name: zookeeper
  namespace: kubeless
spec:
  ports:
  - name: client
    port: 2181
  selector:
    kubeless: zookeeper
---
apiVersion: apps/v1beta1
kind: StatefulSet
metadata:
  name: zoo
  namespace: kubeless
spec:
  serviceName: zoo
  template:
    metadata:
      labels:
        kubeless: zookeeper
    spec:
      containers:
      - env:
        - name: ZOO_SERVERS
          value: server.1=zoo-0.zoo:2888:3888:participant
        - name: ALLOW_ANONYMOUS_LOGIN
          value: "yes"
        image: bitnami/zookeeper@sha256:f66625a8a25070bee18fddf42319ec58f0c49c376b19a5eb252e6a4814f07123
        imagePullPolicy: IfNotPresent
        name: zookeeper
        ports:
        - containerPort: 2181
          name: client
        - containerPort: 2888
          name: peer
        - containerPort: 3888
          name: leader-election
        volumeMounts:
        - mountPath: /bitnami/zookeeper
          name: zookeeper
      volumes:
      - name: zookeeper
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  labels:
    kubeless: controller
  name: kubeless-controller
  namespace: kubeless
spec:
  selector:
    matchLabels:
      kubeless: controller
  template:
    metadata:
      labels:
        kubeless: controller
    spec:
      containers:
      - image: bitnami/kubeless-controller:latest
        imagePullPolicy: IfNotPresent
        name: kubeless-controller
      serviceAccountName: controller-acct
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: controller-acct
  namespace: kubeless
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: kubeless-controller-deployer
rules:
- apiGroups:
  - ""
  resources:
  - services
  - configmaps
  verbs:
  - create
  - get
  - delete
  - list
  - update
  - patch
- apiGroups:
  - apps
  - extensions
  resources:
  - deployments
  verbs:
  - create
  - get
  - delete
  - list
  - update
  - patch
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - list
  - delete
- apiGroups:
  - k8s.io
  resources:
  - functions
  verbs:
  - get
  - list
  - watch
---
apiVersion: apiextensions.k8s.io/v1beta1
description: Kubernetes Native Serverless Framework
kind: CustomResourceDefinition
metadata:
  name: functions.k8s.io
spec:
  group: k8s.io
  names:
    kind: Function
    plural: functions
    singular: function
  scope: Namespaced
  version: v1
---
apiVersion: v1
kind: Service
metadata:
  name: broker
  namespace: kubeless
spec:
  clusterIP: None
  ports:
  - port: 9092
  selector:
    kubeless: kafka
---
apiVersion: apps/v1beta1
kind: StatefulSet
metadata:
  name: kafka
  namespace: kubeless
spec:
  serviceName: broker
  template:
    metadata:
      labels:
        kubeless: kafka
    spec:
      containers:
      - env:
        - name: KAFKA_ADVERTISED_HOST_NAME
          value: broker.kubeless
        - name: KAFKA_ADVERTISED_PORT
          value: "9092"
        - name: KAFKA_PORT
          value: "9092"
        - name: KAFKA_ZOOKEEPER_CONNECT
          value: zookeeper.kubeless:2181
        - name: ALLOW_PLAINTEXT_LISTENER
          value: "yes"
        image: bitnami/kafka@sha256:ef0b1332408c0361d457852622d3a180f3609b9d98f1a85a9a809adaecfe9b52
        imagePullPolicy: IfNotPresent
        livenessProbe:
          initialDelaySeconds: 30
          tcpSocket:
            port: 9092
        name: broker
        ports:
        - containerPort: 9092
        volumeMounts:
        - mountPath: /bitnami/kafka/data
          name: datadir
  volumeClaimTemplates:
  - metadata:
      name: datadir
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: zoo
  namespace: kubeless
spec:
  clusterIP: None
  ports:
  - name: peer
    port: 9092
  - name: leader-election
    port: 3888
  selector:
    kubeless: zookeeper
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  annotations: {}
  labels:
    app: mongodb
    name: mongodb
  name: mongodb
  namespace: kubeapps
spec:
  minReadySeconds: 30
  replicas: 1
  selector:
    matchLabels:
      app: mongodb
      name: mongodb
  strategy:
    rollingUpdate:
      maxSurge: 0
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      annotations: {}
      labels:
        app: mongodb
        name: mongodb
    spec:
      containers:
      - args: []
        env:
        - name: MONGODB_DATABASE
          value: ""
        - name: MONGODB_PASSWORD
          valueFrom:
            secretKeyRef:
              key: mongodb-password
              name: mongodb
        - name: MONGODB_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              key: mongodb-root-password
              name: mongodb
        - name: MONGODB_USERNAME
          value: ""
        image: bitnami/mongodb:3.4.9-r1
        livenessProbe:
          exec:
            command:
            - mongo
            - --eval
            - db.adminCommand('ping')
          initialDelaySeconds: 30
          timeoutSeconds: 5
        name: mongodb
        ports:
        - containerPort: 27017
          name: mongodb
        readinessProbe:
          exec:
            command:
            - mongo
            - --eval
            - db.adminCommand('ping')
          initialDelaySeconds: 5
          timeoutSeconds: 1
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
        stdin: false
        tty: false
        volumeMounts:
        - mountPath: /bitnami/mongodb
          name: data
      imagePullSecrets: []
      initContainers: []
      terminationGracePeriodSeconds: 30
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: mongodb-data
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  annotations: {}
  labels:
    name: mongodb-data
  name: mongodb-data
  namespace: kubeapps
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 8Gi
---
apiVersion: v1
kind: Service
metadata:
  annotations: {}
  labels:
    name: mongodb
  name: mongodb
  namespace: kubeapps
spec:
  ports:
  - port: 27017
    targetPort: mongodb
  selector:
    app: mongodb
    name: mongodb
  type: ClusterIP
---
apiVersion: v1
kind: Namespace
metadata:
  annotations: {}
  labels:
    name: kubeapps
  name: kubeapps
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sealed-secrets-controller
  namespace: kube-system
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: sealed-secrets-controller
  namespace: kube-system
spec:
  template:
    metadata:
      labels:
        name: sealed-secrets-controller
    spec:
      containers:
      - command:
        - controller
        image: quay.io/bitnami/sealed-secrets-controller:v0.5.1
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
        name: sealed-secrets-controller
        ports:
        - containerPort: 8080
          name: http
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8080
        securityContext:
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 1001
      serviceAccountName: sealed-secrets-controller
---
apiVersion: v1
kind: Service
metadata:
  name: sealed-secrets-controller
  namespace: kube-system
spec:
  ports:
  - port: 8080
  selector:
    name: sealed-secrets-controller
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: sealed-secrets-controller
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: sealed-secrets-key-admin
subjects:
- apiGroup: ""
  kind: ServiceAccount
  name: sealed-secrets-controller
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: sealed-secrets-key-admin
  namespace: kube-system
rules:
- apiGroups:
  - ""
  resourceNames:
  - sealed-secrets-key
  resources:
  - secrets
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - create
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: sealed-secrets-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: secrets-unsealer
subjects:
- apiGroup: ""
  kind: ServiceAccount
  name: sealed-secrets-controller
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: secrets-unsealer
rules:
- apiGroups:
  - bitnami.com
  resources:
  - sealedsecrets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - create
  - update
  - delete
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: sealedsecrets.bitnami.com
spec:
  group: bitnami.com
  names:
    kind: SealedSecret
    listKind: SealedSecretList
    plural: sealedsecrets
    singular: sealedsecret
  scope: Namespaced
  validation:
    openAPIV3Schema:
      $schema: http://json-schema.org/draft-04/schema#
      description: A sealed (encrypted) Secret
      properties:
        spec:
          properties:
            data:
              pattern: ^[^A-Za-z0-9+/=]*$
              type: string
          type: object
      type: object
  version: v1alpha1
`
)

var (
	// VERSION will be overwritten automatically by the build system
	VERSION = "devel"
)

// RootCmd is the root of cobra subcommand tree
var RootCmd = &cobra.Command{
	Use:   "kubeapps",
	Short: "Kubeapps Installer manages to install Kubeapps components to your cluster",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStderr()
		logrus.SetOutput(out)
		return nil
	},
}

func bindFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("namespace", "", api.NamespaceDefault, "Specify namespace for the Kubeapps components")
}

func parseObjects(manifest string) ([]*unstructured.Unstructured, error) {
	r := strings.NewReader(manifest)
	decoder := yaml.NewYAMLReader(bufio.NewReader(r))
	ret := []runtime.Object{}
	for {
		bytes, err := decoder.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if len(bytes) == 0 {
			continue
		}
		jsondata, err := yaml.ToJSON(bytes)
		if err != nil {
			return nil, err
		}
		obj, _, err := unstructured.UnstructuredJSONScheme.Decode(jsondata, nil, nil)
		if err != nil {
			return nil, err
		}
		ret = append(ret, obj)
	}

	return utils.FlattenToV1(ret), nil
}

func restClientPool() (dynamic.ClientPool, discovery.DiscoveryInterface, error) {
	conf, err := buildOutOfClusterConfig()
	if err != nil {
		return nil, nil, err
	}

	disco, err := discovery.NewDiscoveryClientForConfig(conf)
	if err != nil {
		return nil, nil, err
	}

	discoCache := utils.NewMemcachedDiscoveryClient(disco)
	mapper := discovery.NewDeferredDiscoveryRESTMapper(discoCache, dynamic.VersionInterfaces)
	pathresolver := dynamic.LegacyAPIPathResolverFunc

	pool := dynamic.NewClientPool(conf, mapper, pathresolver)
	return pool, discoCache, nil
}

func buildOutOfClusterConfig() (*rest.Config, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		home, err := getHome()
		if err != nil {
			return nil, err
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

func getHome() (string, error) {
	home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	if home == "" {
		for _, h := range []string{"HOME", "USERPROFILE"} {
			if home = os.Getenv(h); home != "" {
				return home, nil
			}
		}
	} else {
		return home, nil
	}

	return "", errors.New("can't get home directory")
}
