# Access Control in Kubeapps

Kubeapps requires users to login with a Kubernetes API token in order to make
requests to the Kubernetes API server as the user. This ensures that a certain
user of Kubeapps is only permitted to view and manage applications that they
have access to (for example, within a specific namespace). If a user does not
have access to a particular resource, Kubeapps will display an error describing
the required roles to access the resource.

If your cluster supports [Token
Authentication](https://kubernetes.io/docs/admin/authentication/) you may login
with the same tokens. Alternatively, you can create Service Accounts for
Kubeapps users. The examples below use a Service Account, as it is the most
common scenario.

## Service Accounts

To create a Service Account for a user "example" in the "default" namespace, run
the following:

```
kubectl create -n default serviceaccount example
```

To get the API token for this Service Account, run the following:

```
kubectl get -n default secret $(kubectl get -n default serviceaccount example -o jsonpath='{.secrets[].name}') -o jsonpath='{.data.token}' | base64 --decode
```

## Assigning Kubeapps User Roles

Kubeapps will install a set of preset Roles and ClusterRoles in your cluster
that you can bind to user or Service Accounts. Each Role and ClusterRole
pertains to a certain operation within Kubeapps. This documentation describes
the roles that should be applied to a user in order to perform operations within
Kubeapps.

### Applications

#### Read access to Applications within a namespace

In order to list and view Applications in a namespace, apply the
`kubeapps-applications-read` ClusterRole in the desired namespace and the
`kubeapps-tiller-state-read`:

```
kubectl create -n default rolebinding example-kubeapps-applications-read \
  --clusterrole=kubeapps-applications-read \
  --serviceaccount default:example
```

Note: This will give read access to all the resources in the default namespace.
If you want to provide fine-grained access to resources consider creating your
own roles.

#### Write access to Applications within a namespace

In order to create, update and delete Applications in a namespace, apply the
`kubeapps-applications-write` ClusterRole in the desired namespace and the
`kubeapps-repositories-read` Role in the `kubeapps` namespace:

```
kubectl create -n default rolebinding example-kubeapps-applications-write \
  --clusterrole=kubeapps-applications-write \
  --serviceaccount default:example
kubectl create -n kubeapps rolebinding example-kubeapps-repositories-read \
  --role=kubeapps-repositories-read \
  --serviceaccount default:example
```

Note: This will give write access to all the resources in the default namespace.
If you want to provide fine-grained access to resources consider creating your
own roles.

### Functions

#### Read access to Functions within a namespace

In order to list and view Functions in a namespace, apply the
`kubeapps-functions-read` ClusterRole in the desired namespace and the
`kubeapps-kubeless-config-read` Role in the `kubeless` namespace:

```
kubectl create -n default rolebinding example-kubeapps-functions-read --clusterrole=kubeapps-functions-read --serviceaccount default:example
kubectl create -n kubeless rolebinding example-kubeapps-kubeless-config-read --role=kubeapps-kubeless-config-read --serviceaccount default:example
```

#### Write access to Functions within a namespace

In order to create, update and delete Functions in a namespace, apply the
`kubeapps-functions-write` ClusterRole in the desired namespace.

```
kubectl create -n default rolebinding example-kubeapps-functions-write --clusterrole=kubeapps-functions-write --serviceaccount default:example
```

### Service Catalog, Service Instances and Bindings

#### Read access to Service Instances and Bindings within a namespaces

Service Brokers, Classes and Plans in the Service Catalog are cluster-scoped
resources, but Service Instances and Bindings can be restricted to a namespace.
Kubeapps defines two roles (`kubeapps-service-catalog-browse` and
`kubeapps-service-catalog-read`) to separate the roles required to view Service
Instances and Bindings so that they can be applied to desired namespaces.

In order to list and view Service Instances in a namespace, apply the
`kubeapps-service-catalog-browse` ClusterRole in all namespaces and the
`kubeapps-service-catalog-read` in the desired namespace.

```
kubectl create clusterrolebinding example-kubeapps-service-catalog-browse --clusterrole=kubeapps-service-catalog-browse --serviceaccount default:example
kubectl create -n default rolebinding example-kubeapps-service-catalog-read --clusterrole=kubeapps-service-catalog-read --serviceaccount default:example
```

#### Write access to Service Instances and Bindings within a namespace

In order to create and delete Service Instances and Bindings in a namespace,
apply the `kubeapps-service-catalog-write` ClusterRole in the desired namespace.

```
kubectl create -n default rolebinding example-kubeapps-service-catalog-write --clusterrole=kubeapps-service-catalog-write --serviceaccount default:example
```

#### Admin access to configure Service Brokers

In order to resync Service Brokers from the Service Brokers Configuration page,
apply the `kubeapps-service-catalog-admin` ClusterRole in all namespaces.

```
kubectl create clusterrolebinding example-kubeapps-service-catalog-admin --clusterrole=kubeapps-service-catalog-admin --serviceaccount default:example
```

### App Repositories

#### Read access to App Repositories

In order to list the configured App Repositories in Kubeapps, apply the
`kubeapps-repositories-read` Role in the kubeapps namespace.

```
kubectl create -n kubeapps rolebinding example-kubeapps-repositories-read --role=kubeapps-repositories-read --serviceaccount default:example
```

#### Write access to App Repositories

In order to create and refresh App Repositories in Kubeapps, apply the
`kubeapps-repositories-write` Role in the kubeapps namespace.

```
kubectl create -n kubeapps rolebinding example-kubeapps-repositories-write --role=kubeapps-repositories-write --serviceaccount default:example
```

### Assigning roles across multiple namespaces

To give permissions in multiple namespaces, simply create the same RoleBindings
in each namespace you want to configure access for. For example, to give the
"example" user permissions to manage Applications in the "example" namespace:

```
kubectl create -n example rolebinding example-kubeapps-applications-write --clusterrole=kubeapps-applications-read --serviceaccount default:example
kubectl create -n example rolebinding example-kubeapps-applications-write --clusterrole=kubeapps-applications-write --serviceaccount default:example
```

Note that there's no need to recreate the RoleBinding in the kubeapps namespace
that is also needed, since that has already been created.

If you want to give access for every namespace, simply create a
ClusterRoleBinding instead of a RoleBinding. For example, to give the "example" user permissions to manage Applications in _any_ namespace:

```
kubectl create clusterrolebinding example-kubeapps-applications-write --clusterrole=kubeapps-applications-read --serviceaccount default:example
kubectl create clusterrolebinding example-kubeapps-applications-write --clusterrole=kubeapps-applications-write --serviceaccount default:example
```

## RBAC rules required by Kubeapps

An up-to-date list of RBAC rules Kubeapps requires can be found [here](/manifests/user-roles.jsonnet).

## Using a cluster-admin user (not recommended)

A simpler way to configure access for Kubeapps would be to give the user
cluster-admin access (effectively disabling RBAC). This is not recommended, but
useful for quick demonstrations or evaluations.

```
kubectl create serviceaccount kubeapps-operator
kubectl create clusterrolebinding kubeapps-operator --clusterrole=cluster-admin --serviceaccount=default:kubeapps-operator
```
