# Kubeapps Helm 3 support

We believe that the transition to Helm 3 can be done in such a fashion that
both the old tiller-proxy and the new Helm 3 components can coexist.
However, we propose that the initial effort not touch tiller-proxy, but create
a completely new component.
The choice between Helm 2 support and Helm 3 support should be made at deploy
time in Helm (e.g. by setting `helm3=true` in Values).

Since Helm 3 has gotten rid of Tiller, it instead provides a client library
that can be used to access all required commands.
We see no reason to implement a proxy for Helm 3, but rather what we call an
agent.
We should update the API version from v1 to v2, to clearly distinguish between
the old and the new subsystems.
That will hopefully help with compatibility later on in case the architecture
changes more.

**Current situation:**
![Current situation](https://user-images.githubusercontent.com/7773090/67413010-ac044e00-f5c0-11e9-93e9-f3cdd1eeaca8.PNG)

**With the new additions:**
![With the new additions](https://user-images.githubusercontent.com/7773090/67413025-b45c8900-f5c0-11e9-8961-67377bc8faad.PNG)

## Authentication

Since Helm 2 built on Tiller, all authentication to the Kubernetes cluster
happened "over there" and Kubeapps did not need to authenticate anything other
than the communication with Tiller.
Now, with Helm 3, all authentication is handled by the Kubernetes API Server
via the credentials provided by kubectl.
We have created a PoC using the Helm Library in Helm 3, and we have found that
it should be relatively straight-forward to implement since both `ca.crt` and
the token are already provided.
Basically all we need to do is to call the `InClusterConfig` method and
overwrite the `BearerToken` field in the received configuration with the token
string we get from the Dashboard.

