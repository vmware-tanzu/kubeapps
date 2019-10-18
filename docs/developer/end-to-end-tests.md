# End-to-end tests in the project

In every CI build, a set of end-to-end tests are run to verify, as much as possible, that the changes don't include regressions from an user point of view. The current end-to-end tests are executed in two steps (or categories):

 - Chart tests
 - Browser tests

These tests are executed by the script [scripts/e2e-test.sh](../../script/e2e-test.sh). This script:

 1. Installs Tiller using a certificate
 2. Installs Kubeapps using the images built during the CI process
 3. Waits for the different deployments to be ready
 4. Execute the Helm tests (see the section below for more details).
 5. Execute the web browser tests (see the section below for more details).

If all of the above succeeded, the control is returned to the CI with the proper exit code.

## Chart tests

Chart tests in the project are defined using the testing functionality [provided by Helm](https://helm.sh/docs/developing_charts/#chart-tests). The goal of these tests is that the chart has been successfully deployed and that the basic functionality for each of the microservices deployed work as expected. Specific functionality tests should be covered by either unit tests or browser tests if needed.

You can find the current chart tests in the [chart folder](../../chart/kubeapps/templates/tests).

## Web Browser tests

Apart from the basic functionality tests run by the chart tests, this project contains web browser test that you can find in the [integration](../../integration) folder.

These tests are based on [Puppeteer](https://github.com/GoogleChrome/puppeteer). Puppeteer is a NodeJS library that provides a high-level API to control Chrome or Chromium (in headless mode by default).

On top of Puppeteer we are using the `jest-puppeteer` module that allow us to execute these tests using the same syntax than in the rest of unit-tests that we have in the project.

The `integration` folder pointed above is self-contained. That means that the different dependencies required to run the browser tests are not included in the default `package.json`. In that folder, it can be found a `Dockerfile` used to generate an image with all the dependencies needed to run the browser tests.

It's possible to run these tests either locally or in a container environment.

### Runing browser tests locally

To run the tests locally you just need to install the required dependencies and set the required environment variables:

```bash
cd integration
yarn install
INTEGRATION_ENTRYPOINT=http://kubeapps.local LOGIN_TOKEN=foo yarn start
```

If anything goes wrong, apart from the logs of the test, you can find the screenshot of the failed test in the folder `reports/screenshots`.

### Running browser tests in a pod

Since the CI environment don't have the required dependencies and to provide a reproducible environment, it's possible to run the browser tests in a Kubernetes pod. To do so, you can spin up an instance running the image `kubeapps/integration-tests`. This image contains all the required dependencies and it waits forever so you can execute commands within it. The goal of this setup is that you can copy the latest tests to the image, run the tests and extract the screenshots in case of failure:

```bash
cd integration
# Deploy the executor pod
kubectl apply -f manifests/executor.yaml
pod=$(kubectl get po -l run=integration -o jsonpath="{.items[0].metadata.name}")
# Copy latest tests
kubectl cp ./use-cases ${pod}:/app/
# Run tests
kubectl exec -it ${pod} -- /bin/sh -c 'INTEGRATION_ENTRYPOINT=http://kubeapps.kubeapps LOGIN_TOKEN=foo yarn start'
# If the tests fail, get report screenshot
kubectl cp ${pod}:/app/reports ./reports
```
