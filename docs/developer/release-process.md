# Kubeapps Releases Developer Guide

The purpose of this document is to guide you through the process of releasing a new version of Kubeapps.

## 1 - Create a new git tag

The first step is to tag the repository master branch tip and push it upstream. It is important to note that the tag name will be used as release name.

```bash
export VERSION_NAME="v1.0.0-beta.1"

git tag ${VERSION_NAME}
git push origin ${VERSION_NAME}
```

This will trigger a build, test and **release** [workflow in our CI](https://circleci.com/gh/kubeapps/workflows).
 
## 2 - Complete the GitHub release notes

Once the release job is finished, you will have a GitHub release draft pre-populated. You still must **add a high level description with the release highlights**. Save the draft and **do not publish it yet**.

## 3 - Bump the chart version

At this point, you will have a new set of published docker images as well as some release notes waiting to be published.

But before, we need to create and merge a PR with a chart version bump in `chart/kubeapps/Chart.yaml` ([example](https://github.com/kubeapps/kubeapps/pull/663/files)). This will trigger another CI job that will publish a new version of the chart pointing to the new Docker images built in the step 1.

## 4 - Publish the GitHub release

Once the chart has been published and the release notes reviewed by a peer, publish the release and we are done!

Don't forget to promote the release in #kubeapps!
