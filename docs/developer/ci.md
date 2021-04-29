# Understanding the CircleCI configuration

Kubeapps leverages CircleCI for running the tests (both unit and integration tests), pushing the images and syncing the chart with the official [Bitnami chart](https://github.com/bitnami/charts/tree/master/bitnami/kubeapps). The following image depicts how a successful workflow looks like (after a push to master).

![CircleCI workflow after pushing to master](../img/ci-workflow-master.png "CircleCI workflow after pushing to master")

The main configuration is located at this [CircleCI config file](../../.circleci/config.yml)). At a glance, it contains:

- **Build conditions**: `build_always`, `build_on_master` and `build_on_tag`. They will be added to each job to determine whether or not it should be executed. Whereas some should always be run, others only make sense when pushing to master or when a new tag has been created.
- **Workflows**: we only use a single workflow named `kubeapps` with multiple jobs.
- **Jobs**: the actual commands that are executed depending on the build conditions.
  - `test_go` (always): it runs every unit test for those projects written in Golang (that is, it runs `make test`) as well as it runs some DB-dependent tests.
  - `test_dashboard` (always): it runs the dashboard linter and unit tests (`yarn lint` and `yarn test`)
  - `test_pinniped_proxy` (always): it runs the Rust unit tests of the pinniped-proxy project (`cargo test`).
  - `test_chart_render` (always): it runs the chart template test defined in the script [`chart-template-test.sh](../../script/chart-template-test.sh).
  - `build_go_images` (always): it builds the CI golang images for `kubeops`, `apprepository-controller`, `asset-syncer` and `assetsvc`.
  - `build_dashboard` (always): it builds the CI node image for `dashboard`.
  - `build_pinniped_proxy` (always): it builds the CI rust image for `pinniped-proxy`.
  - `local_e2e_tests` (always): it runs locally (i.e., inside the CircleCI environment) the e2e tests. Please refer to the [e2e tests documentation](./end-to-end-tests.md) for further information. In this job, before executing the script [`script/e2e-test.sh](../../script/e2e-test.sh), the proper environment is created. Namely:
    - Install the required binaries (kind, kubectl, mkcert, helm).
    - Spin up two Kind clusters.
    - Load the CI images into the cluster.
    - Run the integration tests.
  - `GKE_x_xx_MASTER` and `GKE_X_XX_LATEST_RELEASE` (on master): there is a job for each [Kubernetes version supported by Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine/docs/release-notes) (GKE). It will run the e2e tests in a GKE cluster (version X.XX) using either the code in `master` or in the latest released version. If a change affecting the UI is pushed to master, the e2e test might fail here. Use a try/catch block to temporarily work around this.
  - `push_images` (on master): the CI images (which have already been built) get re-tagged and pushed to the `kubeapps` account.
  - `release` (on tag): it creates a GitHub release based on the current tag by executing the script [script/create_release.sh](../../script/create_release.sh).

Note that this process is independent of the release of the official Bitnami images and chart. These Bitnami images will be created according to their internal process (so the Golang, Node or Rust versions we define here are not used by them. Manual coordination is expected here if a major version bump happens to occur).

Additionally, currently is the Kubeapps team who is in charge of sending a PR to the [chart repository](https://github.com/bitnami/charts/tree/master/bitnami/kubeapps) each time a new chart version is pushed to the main branch.
