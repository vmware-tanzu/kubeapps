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

In order to list and view Applications in a namespace, first we will create a `ClusterRole` with read-access to **all** the possible resources. In case you want
to limit this access, create a custom cluster role or use one of the [default ones](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#user-facing-roles). Then we will bind that cluster role to our service account.

```
kubectl apply -f https://raw.githubusercontent.com/kubeapps/kubeapps/master/docs/user/manifests/kubeapps-applications-read.yaml
kubectl create -n default rolebinding example-view \
  --clusterrole=kubeapps-applications-read \
  --serviceaccount default:example
```

#### Write access to Applications within a namespace

In order to create, update and delete Applications in a namespace, apply the
`edit` ClusterRole in the desired namespace. The `edit` ClusterRole should be
available in most Kubernetes distributions, you can find more information about
that role
[here](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#user-facing-roles).

```
kubectl create -n default rolebinding example-edit \
  --clusterrole=edit \
  --serviceaccount default:example
```

### Service Catalog, Service Instances and Bindings

#### Read access to Service Instances and Bindings within a namespaces

Service Brokers, Classes and Plans in the Service Catalog are cluster-scoped
resources, but Service Instances and Bindings can be restricted to a namespace.
We'll need to define two roles (`kubeapps-service-catalog-browse` and
`kubeapps-service-catalog-read`) to separate the roles required to view Service
Instances and Bindings so that they can be applied to desired namespaces.

In order to list and view Service Instances in a namespace, we'll create the
`kubeapps-service-catalog-browse` ClusterRole in all namespaces and the
`kubeapps-service-catalog-read` in the desired namespace.

```
kubectl apply -f https://raw.githubusercontent.com/kubeapps/kubeapps/master/docs/user/manifests/kubeapps-service-catalog-browse.yaml
kubectl create clusterrolebinding example-kubeapps-service-catalog-browse --clusterrole=kubeapps-service-catalog-browse --serviceaccount default:example

kubectl apply -f https://raw.githubusercontent.com/kubeapps/kubeapps/master/docs/user/manifests/kubeapps-service-catalog-read.yaml
kubectl create -n default rolebinding example-kubeapps-service-catalog-read --clusterrole=kubeapps-service-catalog-read --serviceaccount default:example
```

#### Write access to Service Instances and Bindings within a namespace

In order to create and delete Service Instances and Bindings in a namespace,
create and bind the `kubeapps-service-catalog-write` ClusterRole in the desired namespace.

```
kubectl apply -f https://raw.githubusercontent.com/kubeapps/kubeapps/master/docs/user/manifests/kubeapps-service-catalog-write.yaml
kubectl create -n default rolebinding example-kubeapps-service-catalog-write --clusterrole=kubeapps-service-catalog-write --serviceaccount default:example
```

#### Admin access to configure Service Brokers

In order to resync Service Brokers from the Service Brokers Configuration page,
create and apply the `kubeapps-service-catalog-admin` ClusterRole in all namespaces.

```
kubectl apply -f https://raw.githubusercontent.com/kubeapps/kubeapps/master/docs/user/manifests/kubeapps-service-catalog-admin.yaml
kubectl create clusterrolebinding example-kubeapps-service-catalog-admin --clusterrole=kubeapps-service-catalog-admin --serviceaccount default:example
```

### App Repositories

#### Read access to App Repositories

In order to list the configured App Repositories in Kubeapps, [bind users/groups Subjects](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#command-line-utilities) to the `$RELEASE_NAME-repositories-read` role in the namespace Kubeapps was installed into by the helm chart.
```
export KUBEAPPS_NAMESPACE=kubeapps
export KUBEAPPS_RELEASE_NAME=kubeapps
kubectl create -n $KUBEAPPS_NAMESPACE rolebinding example-kubeapps-repositories-read \
  --role=$KUBEAPPS_RELEASE_NAME-repositories-read \
  --serviceaccount default:example
```

#### Write access to App Repositories

Likewise to the read access bind users/group Subjects to the
`$KUBEAPPS_RELEASE_NAME-repositories-write` Role in the namespace Kubeapps is installed in
for users to create and refresh App Repositories in Kubeapps

```
export KUBEAPPS_NAMESPACE=kubeapps
export KUBEAPPS_RELEASE_NAME=kubeapps
kubectl create -n $KUBEAPPS_NAMESPACE rolebinding example-kubeapps-repositories-write \
  --role=$KUBEAPPS_RELEASE_NAME-repositories-write \
  --serviceaccount default:example
```

### Assigning roles across multiple namespaces

To give permissions in multiple namespaces, simply create the same RoleBindings
in each namespace you want to configure access for. For example, to give the
"example" user permissions to manage Applications in the "example" namespace:

```
kubectl create -n example rolebinding example-kubeapps-applications-write --clusterrole=kubeapps-applications-read --serviceaccount default:example
kubectl create -n example rolebinding example-kubeapps-applications-write --clusterrole=kubeapps-applications-write --serviceaccount default:example
```

Note that there's no need to recreate the RoleBinding in the namespace Kubeapps
is installed in that is also needed, since that has already been created.

If you want to give access for every namespace, simply create a
ClusterRoleBinding instead of a RoleBinding. For example, to give the "example" user permissions to manage Applications in _any_ namespace:

```
kubectl create clusterrolebinding example-kubeapps-applications-write --clusterrole=kubeapps-applications-read --serviceaccount default:example
kubectl create clusterrolebinding example-kubeapps-applications-write --clusterrole=kubeapps-applications-write --serviceaccount default:example
```

## Using a cluster-admin user (not recommended)

A simpler way to configure access for Kubeapps would be to give the user
cluster-admin access (effectively disabling RBAC). This is not recommended, but
useful for quick demonstrations or evaluations.

```
kubectl create serviceaccount kubeapps-operator
kubectl create clusterrolebinding kubeapps-operator --clusterrole=cluster-admin --serviceaccount=default:kubeapps-operator
```
