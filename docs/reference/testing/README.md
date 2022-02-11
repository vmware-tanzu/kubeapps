## Kubeapps testing guides

### Setup Kubeapps for testing 

This guide explains how to setup your environment to test Kubeapps integration with other services.

Please refer to [Setup Kubeapps testing enviroment](./testing-environment.md)


### CI configuration

Kubeapps leverages CircleCI for running the tests (both unit and integration tests), pushing the images and syncing the chart with the official [Bitnami chart](https://github.com/bitnami/charts/tree/master/bitnami/kubeapps).

Please refer to [Understanding the CircleCI configuration](./ci.md)


### End-to-end test

Kubeapps includes a set of end-to-end tests that are run to verify, as much as possible, that the changes don't include regressions from a user point of view. 

Please refer to [end-to-end testing guide](./end-to-end-tests.md) to know more about end-to-end tests in Kubeapps.

