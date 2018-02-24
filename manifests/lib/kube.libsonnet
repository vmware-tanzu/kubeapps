// Generic library of Kubernetes objects
// TODO: Expand this to include all API objects.
//
// Should probably fill out all the defaults here too, so jsonnet can
// reference them.  In addition, jsonnet validation is more useful
// (client-side, and gives better line information).

{
  // Returns array of values from given object.  Does not include hidden fields.
  objectValues(o):: [o[field] for field in std.objectFields(o)],

  // Returns array of [key, value] pairs from given object.  Does not include hidden fields.
  objectItems(o):: [[k, o[k]] for k in std.objectFields(o)],

  // Replace all occurrences of `_` with `-`.
  hyphenate(s):: std.join("-", std.split(s, "_")),

  // Convert {foo: {a: b}} to [{name: foo, a: b}]
  mapToNamedList(o):: [{ name: $.hyphenate(n) } + o[n] for n in std.objectFields(o)],

  // Convert from SI unit suffixes to regular number
  siToNum(n):: (
    local convert =
      if std.endsWith(n, "m") then [1, 0.001]
      else if std.endsWith(n, "K") then [1, 1e3]
      else if std.endsWith(n, "M") then [1, 1e6]
      else if std.endsWith(n, "G") then [1, 1e9]
      else if std.endsWith(n, "T") then [1, 1e12]
      else if std.endsWith(n, "P") then [1, 1e15]
      else if std.endsWith(n, "E") then [1, 1e18]
      else if std.endsWith(n, "Ki") then [2, std.pow(2, 10)]
      else if std.endsWith(n, "Mi") then [2, std.pow(2, 20)]
      else if std.endsWith(n, "Gi") then [2, std.pow(2, 30)]
      else if std.endsWith(n, "Ti") then [2, std.pow(2, 40)]
      else if std.endsWith(n, "Pi") then [2, std.pow(2, 50)]
      else if std.endsWith(n, "Ei") then [2, std.pow(2, 60)]
      else error "Unknown numerical suffix in " + n;
    local n_len = std.length(n);
    std.parseInt(std.substr(n, 0, n_len - convert[0])) * convert[1]
  ),

  _Object(apiVersion, kind, name):: {
    apiVersion: apiVersion,
    kind: kind,
    metadata: {
      name: name,
      labels: { name: name },
      annotations: {},
    },
  },

  List(): {
    apiVersion: "v1",
    kind: "List",
    items_:: {},
    items: $.objectValues(self.items_),
  },

  Namespace(name): $._Object("v1", "Namespace", name) {
  },

  Endpoints(name): $._Object("v1", "Endpoints", name) {
    Ip(addr):: { ip: addr },
    Port(p):: { port: p },

    subsets: [],
  },

  Service(name): $._Object("v1", "Service", name) {
    local service = self,

    target_pod:: error "service target_pod required",
    port:: self.target_pod.spec.containers[0].ports[0].containerPort,

    // Helpers that format host:port in various ways
    http_url:: "http://%s.%s:%s/" % [
      self.metadata.name,
      self.metadata.namespace,
      self.spec.ports[0].port,
    ],
    proxy_urlpath:: "/api/v1/proxy/namespaces/%s/services/%s/" % [
      self.metadata.namespace,
      self.metadata.name,
    ],
    // Useful in Ingress rules
    name_port:: {
      serviceName: service.metadata.name,
      servicePort: service.spec.ports[0].port,
    },

    spec: {
      selector: service.target_pod.metadata.labels,
      ports: [
        {
          port: service.port,
          targetPort: service.target_pod.spec.containers[0].ports[0].name,
        },
      ],
      type: "ClusterIP",
    },
  },

  PersistentVolume(name): $._Object("v1", "PersistentVolume", name) {
    spec: {},
  },

  // TODO: This is a terrible name
  PersistentVolumeClaimVolume(pvc): {
    persistentVolumeClaim: { claimName: pvc.metadata.name },
  },

  StorageClass(name): $._Object("storage.k8s.io/v1beta1", "StorageClass", name) {
    provisioner: error "provisioner required",
  },

  PersistentVolumeClaim(name): $._Object("v1", "PersistentVolumeClaim", name) {
    local pvc = self,

    storageClass:: null,
    storage:: error "storage required",

    metadata+: if pvc.storageClass != null then {
      annotations+: {
        "volume.beta.kubernetes.io/storage-class": pvc.storageClass,
      },
    } else {},

    spec: {
      resources: {
        requests: {
          storage: pvc.storage,
        },
      },
      accessModes: ["ReadWriteOnce"],
    },
  },

  Container(name): {
    name: name,
    image: error "container image value required",

    envList(map):: [
      if std.type(map[x]) == "object" then { name: x, valueFrom: map[x] } else { name: x, value: map[x] }
      for x in std.objectFields(map)
    ],

    env_:: {},
    env: self.envList(self.env_),

    args_:: {},
    args: ["--%s=%s" % kv for kv in $.objectItems(self.args_)],

    ports_:: {},
    ports: $.mapToNamedList(self.ports_),

    volumeMounts_:: {},
    volumeMounts: $.mapToNamedList(self.volumeMounts_),

    stdin: false,
    tty: false,
    assert !self.tty || self.stdin : "tty=true requires stdin=true",
  },

  Pod(name): $._Object("v1", "Pod", name) {
    spec: $.PodSpec,
  },

  PodSpec: {
    // The 'first' container is used in various defaults in k8s.
    default_container:: std.objectFields(self.containers_)[0],
    containers_:: {},

    containers: [{ name: $.hyphenate(name) } + self.containers_[name] for name in [self.default_container] + [n for n in std.objectFields(self.containers_) if n != self.default_container]],

    volumes_:: {},
    volumes: $.mapToNamedList(self.volumes_),

    imagePullSecrets: [],

    terminationGracePeriodSeconds: 30,

    assert std.length(self.containers) > 0 : "must have at least one container",
  },

  EmptyDirVolume(): {
    emptyDir: {},
  },

  HostPathVolume(path): {
    hostPath: { path: path },
  },

  GitRepoVolume(repository, revision): {
    gitRepo: {
      repository: repository,

      // "master" is possible, but should be avoided for production
      revision: revision,
    },
  },

  SecretVolume(secret): {
    secret: { secretName: secret.metadata.name },
  },

  ConfigMapVolume(configmap): {
    configMap: { name: configmap.metadata.name },
  },

  ConfigMap(name): $._Object("v1", "ConfigMap", name) {
    data: {},

    // I keep thinking data values can be any JSON type.  This check
    // will remind me that they must be strings :(
    local nonstrings = [
      k
      for k in std.objectFields(self.data)
      if std.type(self.data[k]) != "string"
    ],
    assert std.length(nonstrings) == 0 : "data contains non-string values: %s" % [nonstrings],
  },

  // subtype of EnvVarSource
  ConfigMapRef(configmap, key): {
    assert std.objectHas(configmap.data, key) : "%s not in configmap.data" % [key],
    configMapKeyRef: {
      name: configmap.metadata.name,
      key: key,
    },
  },

  Secret(name): $._Object("v1", "Secret", name) {
    local secret = self,

    type: "Opaque",
    data_:: {},
    data: { [k]: std.base64(secret.data_[k]) for k in std.objectFields(secret.data_) },
  },

  // subtype of EnvVarSource
  SecretKeyRef(secret, key): {
    assert std.objectHas(secret.data, key) : "%s not in secret.data" % [key],
    secretKeyRef: {
      name: secret.metadata.name,
      key: key,
    },
  },

  // subtype of EnvVarSource
  FieldRef(key): {
    fieldRef: {
      apiVersion: "v1",
      fieldPath: key,
    },
  },

  // subtype of EnvVarSource
  ResourceFieldRef(key): {
    resourceFieldRef: {
      resource: key,
      divisor_:: 1,
      divisor: std.toString(self.divisor_),
    },
  },

  Deployment(name): $._Object("extensions/v1beta1", "Deployment", name) {
    local deployment = self,

    spec: {
      template: {
        spec: $.PodSpec,
        metadata: {
          labels: deployment.metadata.labels,
          annotations: {},
        },
      },

      strategy: {
        type: "RollingUpdate",

        local pvcs = [
          v
          for v in deployment.spec.template.spec.volumes
          if std.objectHas(v, "persistentVolumeClaim")
        ],
        local is_stateless = std.length(pvcs) == 0,

        // Apps trying to maintain a majority quorum or similar will
        // want to tune these carefully.
        // NB: Upstream default is surge=1 unavail=1
        rollingUpdate: if is_stateless then {
          maxSurge: "25%",  // rounds up
          maxUnavailable: "25%",  // rounds down
        } else {
          // Poor-man's StatelessSet.  Useful mostly with replicas=1.
          maxSurge: 0,
          maxUnavailable: 1,
        },
      },

      // NB: Upstream default is 0
      minReadySeconds: 30,

      // NB: Regular k8s default is to keep all revisions
      revisionHistoryLimit: 10,

      replicas: 1,
      assert self.replicas >= 1,
    },
  },

  CrossVersionObjectReference(target): {
    apiVersion: target.apiVersion,
    kind: target.kind,
    name: target.metadata.name,
  },

  HorizontalPodAutoscaler(name): $._Object("autoscaling/v1", "HorizontalPodAutoscaler", name) {
    local hpa = self,

    target:: error "target required",

    spec: {
      scaleTargetRef: $.CrossVersionObjectReference(hpa.target),

      minReplicas: hpa.target.spec.replicas,
      maxReplicas: error "maxReplicas required",

      assert self.maxReplicas >= self.minReplicas,
    },
  },

  StatefulSet(name): $._Object("apps/v1beta1", "StatefulSet", name) {
    local sset = self,

    spec: {
      serviceName: name,

      template: {
        spec: $.PodSpec,
        metadata: {
          labels: sset.metadata.labels,
          annotations: {},
        },
      },

      volumeClaimTemplates_:: {},
      volumeClaimTemplates: [$.PersistentVolumeClaim($.hyphenate(kv[0])) + kv[1] for kv in $.objectItems(self.volumeClaimTemplates_)],

      replicas: 1,
      assert self.replicas >= 1,
    },
  },

  Job(name): $._Object("batch/v1", "Job", name) {
    local job = self,

    spec: {
      template: {
        spec: $.PodSpec {
          restartPolicy: "OnFailure",
        },
        metadata: {
          labels: job.metadata.labels,
          annotations: {},
        },
      },

      completions: 1,
      parallelism: 1,
    },
  },

  DaemonSet(name): $._Object("extensions/v1beta1", "DaemonSet", name) {
    local ds = self,
    spec: {
      template: {
        metadata: {
          labels: ds.metadata.labels,
          annotations: {},
        },
        spec: $.PodSpec,
      },
    },
  },

  Ingress(name): $._Object("extensions/v1beta1", "Ingress", name) {
    spec: {},
  },

  ThirdPartyResource(name): $._Object("extensions/v1beta1", "ThirdPartyResource", name) {
    versions_:: [],
    versions: [{ name: n } for n in self.versions_],
  },

  CustomResourceDefinition(name): $._Object("apiextensions.k8s.io/v1beta1", "CustomResourceDefinition", name) {
    spec: {},
  },

  ServiceAccount(name): $._Object("v1", "ServiceAccount", name) {
  },

  Role(name): $._Object("rbac.authorization.k8s.io/v1beta1", "Role", name) {
    rules: [],
  },

  ClusterRole(name): $.Role(name) {
    kind: "ClusterRole",
  },

  Group(name): {
    kind: "Group",
    name: name,
    apiGroup: "rbac.authorization.k8s.io",
  },

  User(name): {
    kind: "User",
    name: name,
    apiGroup: "rbac.authorization.k8s.io",
  },

  RoleBinding(name): $._Object("rbac.authorization.k8s.io/v1beta1", "RoleBinding", name) {
    local rb = self,

    subjects_:: [],
    subjects: [{
      kind: o.kind,
      namespace: o.metadata.namespace,
      name: o.metadata.name,
    } for o in self.subjects_],

    roleRef_:: error "roleRef is required",
    roleRef: {
      apiGroup: "rbac.authorization.k8s.io",
      kind: rb.roleRef_.kind,
      name: rb.roleRef_.metadata.name,
    },
  },

  ClusterRoleBinding(name): $.RoleBinding(name) {
    kind: "ClusterRoleBinding",
  },
}
