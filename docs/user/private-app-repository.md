# Using a Private Repository with Kubeapps

It is possible to use a private Helm repository to store your own Helm charts and deploy them using Kubeapps. In this guide we will show how you can do that with some of the solutions available right now:

- [ChartMuseum](#chartmuseum)
- [Artifactory](#artifactory) (Pro)

## ChartMuseum

[ChartMuseum](https://chartmuseum.com) is an open-source Helm Chart Repository written in Go (Golang), with support for cloud storage backends, including Google Cloud Storage, Amazon S3, Microsoft Azure Blob Storage, Alibaba Cloud OSS Storage and OpenStack Object Storage.

To use ChartMuseum with Kubeapps, the first thing is to deploy its Helm chart that we can find in the `stable` repository:

<img src="../img/chartmuseum-chart.png" alt="ChartMuseum Chart" width="600px">

In the deployment form we should change at least two things:

- `env.open.DISABLE_API`: We should set this value to `false` so we can use the ChartMuseum API to push new charts.
- `persistence.enabled`: We will put this value to `true` to enable persistence for the charts we store. Note that this will create a [Kubernetes Persistent Volume Claim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#lifecycle-of-a-volume-and-claim) so depending on your Kubernetes provider you may need to manually allocate the required Persistent Volume to satisfy the claim. Some Kubernetes providers will automatically create PVs for you so setting this value to `true` will be enough.

<img src="../img/chartmuseum-deploy-form.png" alt="ChartMuseum Deploy Form" width="600px">

### ChartMuseum: Upload a Chart

Once the chart is deployed you will be able to upload a chart. In one terminal open a port-forward tunnel to the application:

```console
$ export POD_NAME=$(kubectl get pods --namespace default -l "app=chartmuseum" -l "release=my-chartrepo" -o jsonpath="{.items[0].metadata.name}")
$ kubectl port-forward $POD_NAME 8080:8080 --namespace default
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
```

And in a different terminal you can push your chart:

```console
$ helm package /kubeapps/chart/kubeapps
Successfully packaged chart and saved it to: /kubeapps/chart/kubeapps/kubeapps-1.0.0.tgz
$ curl --data-binary "@kubeapps-1.0.0.tgz" http://localhost:8080/api/charts
{"saved":true}
```

### ChartMuseum: Configure the repository in Kubeapps

For adding your private repository go to `Configuration > App Repositories` in Kubeapps and click on "Add App Repository". You will need to add your repository using as URL the service of the application. This will be `<release_name>-chartmuseum.<namespace>:8080`:

<img src="../img/chartmuseum-repository.png" alt="ChartMuseum App Repository" width="600px">

Once you create the repository you can click in the link for the specific repository and you will be able to deploy your own application using Kubeapps.

### ChartMuseum: Authentication/Authorization

It is possible to configure ChartMuseum to use authentication with two different mechanism:

- Using HTTP [basic authentication](https://chartmuseum.com/docs/#basic-auth) (user/password). To use this feature, it's needed to:
  - Specify the parameters `secret.AUTH_USER` and `secret.AUTH_PASS` when deploying the ChartMuseum.
  - Change the URL of the App Repository in Kubeapps to include the credentials: `http://<user>:<password>@<release_name>.<namespace>:8080`
- Using a [JWT token](https://github.com/chartmuseum/auth-server-example). Once you obtain a valid token you can set it in the Authorization Header field of the App Repository form.

## Artifactory

JFrog Artifactory is a Repository Manager supporting all major packaging formats, build tools and CI servers.

> **Note**: In order to use the Helm repository feature, it's necessary to use an Artifactory Pro account.

To use Artifactory with Kubeapps the first thing is adding the JFrog repository to Kubeapps. Go to `Configuration > App Repositories` and add their repository:

<img src="../img/jfrog-repository.png" alt="JFrog repository" width="400px">

Then click on the JFrog repository and deploy Artifactory. For detailed installation instructions, check its [README](https://github.com/jfrog/charts/tree/master/stable/artifactory). If you don't have any further requirement, the values by default will work.

When deployed, during the first login, select "Helm" to initialize a repository:

<img src="../img/jfrog-wizard.png" alt="JFrog repository" width="600px">

By default, Artifactory creates a chart repository called `helm`. That is the one you can use to store your applications.

### Artifactory: Upload a chart

First, it's needed to obtain the user and password of the Helm repository. To obtain it, click in the `helm` repository and, in the `Set Me Up` menu, introduce your password. After that you will be able to see the user and password.

Once you have done that, you will be able to upload a chart:

```
$ curl -u{USER}:{PASSWORD} -T /path/to/chart.tgz "http://{REPO_URL}/artifactory/helm/"
```

### Artifactory: Configure the repository in Kubeapps

To be able able to access private charts with Kubeapps first you need to generate a token. You can do that with the Artifactory API:

```
curl -u{USER}:{PASSWORD} -XPOST "http://{REPO_URL}/artifactory/api/security/token?expires_in=0" -d "username=kubeapps" -d "scope=member-of-groups:readers"
{
  "scope" : "member-of-groups:readers api:*",
  "access_token" : "TOKEN CONTENT",
  "token_type" : "Bearer"
}
```

With that you have created a token with read-only permissions. Go to the `Configuration > App Repositories` menu and add your personal repository:

<img src="../img/jfrog-custom-repo.png" alt="JFrog custom repository" width="400px">

After submitting the repository, you will be able to click in the new repository and see the chart uploaded in the previous step.
