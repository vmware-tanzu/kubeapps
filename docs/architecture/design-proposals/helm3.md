# Kubeapps Helm 3 support

We believe that the transition to Helm 3 can be done in such a fashion that both
the old tiller-proxy and the new Helm 3 components coexists. However, we propose
that the initial effort do not touch the tiller-proxy, but create a completely
new component. Choosing between Helm 2 support and Helm 3 support should be done
at deploy time in Helm (e.g. by setting `helm3=true` in Values).

Since Helm 3 have gotten rid of Tiller, it instead provides a client library
that can be used to access all required commands. We see no reason to implement
a proxy for Helm 3, but rather a client. We should update the API version from
v1 to v2, to clearly distinguish between the old and the new subsystems. That
will hopefully help with compatibility later on in case the architecture changes
more.

## Authentication

Since helm 2 built on tiller, all authentication to the k8s cluster happened
“over there” and Kubeapps did not need to authenticate anything else than the
communication with Tiller. Now, with Helm 3, all authentication is handled by
the Kubernetes client (kubectl). We have created a PoC using the Kubernetes
client library, and we have found that it should be relatively straight-forward
to implement since both ca.crt and the token are already provided. Basically all
we need to do is to create a mock file instead of `token` and overwrite the
BearerToken field in the configuration with the token string we get from the
Dashboard.

