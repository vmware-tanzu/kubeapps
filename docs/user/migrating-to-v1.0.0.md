# Migration to v1.0.0

This release includes a breaking change **if you are upgrading from a version prior to `v1.0.0-beta.2`**. In the version `v1.0.0-beta.2` we introduced a change in MongoDB credentials generation that requires a migration hook that is only present from the version `v1.0.0-beta.2` to `v1.0.0-beta.4`. If you are upgrading Kubeapps from an older version please upgrade first to `v1.0.0-beta.4` to run the necessary migration job.
