# Securing Kubeapps installation

In this guide we will explain how to secure the installation of Kubeapps in a multi-tenant cluster. Following these steps are only necessary if different people with different permissions have access to the same cluster. Generic instructions to secure Helm can be found [here](https://github.com/kubernetes/helm/blob/master/docs/securing_installation.md).

The main goal is to secure the access to [Tiller](https://github.com/kubernetes/helm/blob/master/docs/securing_installation.md) (Helm server-side component). Tiller has access to create or delete any resource in the cluster so we should be careful on how we expose the functionality it provides.

Kubeapps deploys by default Tiller configuring it to listen for connections only in `localhost`. Note that one particularity of Tiller is that even if it is only listening in `localhost`, request made with the `helm` tool can still contact it because, under the hoods, what the client does is opening a [`port-forward`](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) session and connect to it to like a local connection. That means that we still need to add additional security measures to prevent unauthorized access.

Since Tiller is only listening in `localhost`, for exposing Tiller functionality, Kubeapps also deploys as a [sidecar container](https://kubernetes.io/blog/2015/06/the-distributed-system-toolkit-patterns/) a [**proxy**](/cmd/tiller-proxy/README.md) that receives all the requests from the Dashboard (or any other client). This component validates the requests checking that the user is allowed to perform the requested operation and finally redirect it to Tiller.

In order to take advantage of Kubeapps security features you will need to configure two things: a **TLS certificate** to control the access to Tiller and [**RBAC roles**](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) to authorize requests.

# Deploy Kubeapps with a TLS certificate

The first step to restrict the access to Tiller is to use a TLS certificate. If we don't do this any user with access to the Pod in which Tiller is running can escalate privileges. We can generate a self-signed certificate using the tool [`openssl`](https://www.openssl.org/source/). The first thing is creating a certificate authority. You will be asked to introduce a pass phrase for that:

```
$ openssl genrsa -des3 -out kubeapps-ca.key 2048
Generating RSA private key, 2048 bit long modulus
.....................+++
...................+++
e is 65537 (0x10001)
Enter pass phrase for kubeapps-ca.key:
Verifying - Enter pass phrase for kubeapps-ca.key:
```

Now we can generate a root certificate with that key. You will be asked for the previous password:

```
$ openssl req -x509 -new -nodes -key kubeapps-ca.key -sha256 -out kubeapps-ca.pem -subj '/CN=localhost'
Enter pass phrase for kubeapps-ca.key:
```

Having a certificate authority we can create a proper certificate and sign it. Again you will be asked to introduce the password from the first step:

```
$ openssl genrsa -out kubeapps.key 2048
Generating RSA private key, 2048 bit long modulus
.............................+++
...................................................................................+++
e is 65537 (0x10001)

$ openssl req -new -key kubeapps.key -out kubeapps.csr -subj '/CN=localhost'

$ openssl x509 -req -in kubeapps.csr \
  -CA kubeapps-ca.pem -CAkey kubeapps-ca.key -CAcreateserial \
  -out kubeapps.crt -sha256
Signature ok
subject=/C=EN
Getting CA Private Key
Enter pass phrase for kubeapps-ca.key:
```

Now that we have all the files needed to use the certificate we just need to deploy (or update) Kubeapps:

```
kubeapps up --tiller-tls \
  --tiller-tls-cert kubeapps.crt \
  --tiller-tls-key kubeapps.key \
  --tls-ca-cert kubeapps-ca.pem
```

The above command will generate a secret in the `kubeapps` namespace to store the certificate we've just generated and it will configure Tiller to be accessible using only that certificate. We can verify that executing the `helm` tool locally:

```
$ helm version --tiller-namespace kubeapps
...
Error: cannot connect to Tiller
$ helm version --tiller-namespace kubeapps --tls \
  --tls-ca-cert kubeapps-ca.pem \
  --tls-cert kubeapps.crt \
  --tls-key kubeapps.key
...
Server: &version.Version{SemVer:"v2.9.1", GitCommit:"20adb27c7c5868466912eebdf6664e7390ebe710", GitTreeState:"clean"}
```

# Enable RBAC

In order to be able to authorize requests from users it is necessary to enable [RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) in the Kubernetes cluster. Some providers have it enabled by default but in some cases you need to set it up explicitly. Check out your provider documentation to know how to enable it. To verify if your cluster has RBAC available you can check if the API group exists:

```
$ kubectl api-versions | grep rbac.authorization
rbac.authorization.k8s.io/v1
```

Once your cluster has RBAC enabled read [this document](/docs/user/access-control.md) to know how to login in Kubeapps using a token that identifies a user account and how you can create users with different permissions.

In a nutshell, Kubeapps authorization validates:

 - When getting a release details, it checks that the user have "read" access to all the components of the release.
 - When creating, upgrading or deleting a release it checks that the user is allowed to create, update or delete all the components contained in the release chart.

For example, if the user account `foo` wants to deploy a chart `bar` that is composed of a `Deployment` and a `Service` it should have enough permissions to create each one of those. In other case it will receive an error message with the missing permissions required to deploy the chart.

# Storing releases as secrets

One final important point is that Kubeapps configures Tiller to store releases information as `Secrets`. The default helm installation uses `ConfigMaps` instead. `ConfigMaps` are usually less protected by RBAC roles which could leak sensitive information like passwords or other credentials. 

If you already have releases stored as `ConfigMaps` (and that is the default for versions of Kubeapps prior to 1.0.0-alpha.5) you will need to migrate them as secrets to be able to see and manage them through Kubeapps. In order to do so there is an auxiliary command in Kubeapps that helps you to do so:

```
$ kubeapps migrate-configmaps-to-secrets
2018/07/09 16:25:44 Migrated foo.v1 as a secret
2018/07/09 16:25:44 Migrated foo.v2 as a secret
2018/07/09 16:25:44 Migrated wp.v1 as a secret
2018/07/09 16:25:44 Done. ConfigMaps are left in the namespace kubeapps to debug possible errors. Please delete them manually
```

As you can see from the example we have automatically migrated all the revisions of the releases `foo` and `wp` to secrets. Once we are sure that everything work as expected we can delete the remaining `ConfigMaps` to avoid security issues:

```
$ kubectl delete configmaps -n kubeapps -l OWNER=TILLER
configmap "foo.v1" deleted
configmap "foo.v2" deleted
configmap "wp.v1" deleted
```
