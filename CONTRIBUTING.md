# Contributing to Kubeapps

Kubeapps maintainers welcome contributions from the community and first want to thank you for taking the time to contribute!

Please familiarize yourself with the [Code of Conduct](https://github.com/vmware/.github/blob/main/CODE_OF_CONDUCT.md) before contributing.

- _CLA: Before you start working with Kubeapps, please read and sign our Contributor License Agreement [CLA](https://cla.vmware.com/cla/1/preview). If you wish to contribute code and you have not signed our contributor license agreement (CLA), our bot will update the issue when you open a Pull Request. For any questions about the CLA process, please refer to our [FAQ](https://cla.vmware.com/faq)._

## Ways to contribute

Kubeapps project welcomes many different types of contributions and not all of them need a Pull request. Contributions may include:

- New features and proposals
- Documentation
- Bug fixes
- Issue Triage
- Answering questions and giving feedback
- Helping to onboard new contributors
- Other related activities

## Getting started

Find information about how to set up the development environment on the [developer guide](./site/content/docs/latest/reference/developer/README.md).

## Contribution Flow

This is a rough outline of what a contributor's workflow looks like:

- Make a fork of the repository within your GitHub account
- Create a topic branch in your fork from where you want to base your work
- Make commits of logical units
- Make sure your commit messages are with the proper format, quality and descriptiveness (see below)
- All commits must be:
  - Signed using GPG (see [Signing commits in GitHub](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits))
  - Signed off with the line `Signed-off-by: <Your-Name> <Your-email>`. See [related GitHub blogpost about signing off](https://github.blog/changelog/2022-06-08-admins-can-require-sign-off-on-web-based-commits/).
    > Note: Signing off on a commit is different than signing a commit, such as with a GPG key.
- Push your changes to the topic branch in your fork
- Create a pull request containing that commit

Kubeapps maintainers team follow the GitHub workflow and you can find more details on the [GitHub flow documentation](https://docs.github.com/en/get-started/quickstart/github-flow).

Before submitting your pull request use the following checklist:

### Pull Request Checklist

1. Check if your code changes will pass both code linting checks and unit tests.
2. Ensure your commit messages are descriptive. Kubeapps follow the conventions on [How to Write a Git Commit Message](http://chris.beams.io/posts/git-commit/). Be sure to include any related GitHub issue references in the commit message. See [GFM syntax](https://guides.github.com/features/mastering-markdown/#GitHub-flavored-markdown) for referencing issues and commits.
3. Check the commits and commits messages and ensure they are free from typos.
4. Make sure all the commits have been properly signed with GPG and contain the signoff.
5. Any pull request which adds a new feature or changes the behavior of any feature which was previously documented should include updates to the documentation. All documentation lives in this repository.

## Reporting Bugs and Creating Issues

For specifics on what to include in your report, please follow the guidelines in the issue and pull request templates when available.

### Issues

Need an idea for a project to get started contributing? Take a look at the [open issues](https://github.com/vmware-tanzu/kubeapps/issues?q=is%3Aopen+is%3Aissue). Also check to see if any open issues are labeled with [`good first issue`](https://github.com/vmware-tanzu/kubeapps/labels/good%20first%20issue) or [`help wanted`](https://github.com/vmware-tanzu/kubeapps/labels/help%20wanted).

When contributing to Kubeapps, please first discuss the change you wish to make via an issue with this repository before making a change.

> Kubeapps distribution is delegated to the official [Bitnami Kubeapps chart](https://github.com/bitnami/charts/tree/master/bitnami/kubeapps) from the separate Bitnami charts repository. PRs and issues related to the official chart should be created in the Bitnami charts repository.

### Bugs

To file a bug report, please first open an [issue](https://github.com/vmware-tanzu/kubeapps/issues/new?assignees=&labels=kind%2Fbug&template=bug_report.md&title=). The project maintainers team will work with you on your bug report.

Once the bug has been validated, a [pull request](https://github.com/vmware-tanzu/kubeapps/compare) can be opened to fix the bug.

For specifics on what to include in your bug report, please follow the guidelines in the issue and pull request templates.

### Features

To suggest a feature, please first open an [issue](https://github.com/vmware-tanzu/kubeapps/issues/new?assignees=&labels=kind%2Ffeature&template=feature_request.md&title=) that will be tagged with [`kind/proposal`](https://github.com/vmware-tanzu/kubeapps/labels/kind%2Fproposal), or create a new [Discussion](https://github.com/vmware-tanzu/kubeapps/discussions/new). The project maintainers will work with you on your feature request.

Once the feature request has been validated, a pull request can be opened to implement the feature.

For specifics on what to include in your feature request, please follow the guidelines in the issue and pull request templates.

## Ask for Help

The best way to reach Kubeapps maintainers team with a question when contributing is to ask on:

- [GitHub Issues](https://github.com/vmware-tanzu/kubeapps/issues)
- [GitHub Discussions](https://github.com/vmware-tanzu/kubeapps/discussions)
- [#kubeapps Slack channel](https://kubernetes.slack.com/messages/kubeapps)

## Additional Resources

New to Kubeapps?

- Start here to learn how to install and use Kubeapps: [Getting started in Kubeapps](./site/content/docs/latest/tutorials/getting-started.md)
- Start here to learn how to develop for Kubeapps components: [Kubeapps Developer guidelines](./site/content/docs/latest/reference/developer/README.md)
- Other more detailed documentation can be found at: [Kubeapps Docs](./site/content/docs/latest/README.md)

## Roadmap

The near-term and mid-term roadmap for the work planned for the project [maintainers](./MAINTAINERS.md) is documented in [ROADMAP.md](./ROADMAP.md).
