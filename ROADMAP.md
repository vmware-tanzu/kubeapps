## **Kubeapps Project Roadmap**

###

**About this document**

This document aims at providing a higher-level picture of the [Kubeapps Project Board](https://github.com/kubeapps/kubeapps/issues), which serves as the source of truth of the current backlog of issues. These tasks are moved in accordance with our issue [triage process](./docs/developer/issue-triage-process.md).

Therefore, this document is intended to be a reference point for Kubeapps users and contributors to understand where the project is heading, and help determine if a contribution might conflict with a longer-term plan.

###

**How to help?**

Discussion on the roadmap can take place in threads under [Issues](https://github.com/kubeapps/kubeapps/issues). Please open and comment on an issue if you want to provide suggestions and feedback to an item in the roadmap.
Please review the roadmap to avoid potential duplicated efforts. Anyways, you can always reach out to us on the [#Kubeapps channel](https://kubernetes.slack.com/messages/kubeapps) on the Kubernetes Slack server.

###

**How to add an item to the roadmap?**

Please [open an issue](https://github.com/kubeapps/kubeapps/issues/new) to track any initiative on the roadmap of Kubeapps (usually driven by new feature requests). We will work with and rely on our community to focus our efforts on improving Kubeapps.

###

**Current Roadmap**

The following table includes the current roadmap for Kubeapps. If you have any questions or you would like to contribute to Kubeapps, please contact us on Slack on the [#Kubeapps channel](https://kubernetes.slack.com/messages/kubeapps) to discuss with our team.
If you don't know where to start, we are always looking for contributors that will help us reduce technical, automation, and documentation debt.

Please take the dates as _best-effort goals_ since our actual priorities and requirements are subject to change based on community feedback, roadblocks encountered, community contributions, etc.
If you depend on a specific item, we encourage you to reach out to us on Slack, or help us deliver that feature by contributing to Kubeapps.

_Last Updated: August 2021_

|              Topic               |                                                                                                                             Description                                                                                                                             |     Timeline      |
| :------------------------------: | :-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------: | :---------------: |
|      Pluggable architecture      |                         Kubeapps moves forward to define a plugin interface as an extensible way to expand Kubeapps to support new packaging formats such as declarative Flux resources, Carvel bundles and other future packaging formats.                         |     Q3 - 2021     |
|      Kubeapps APIs service       |                                                                                         Consistent API for interacting with packages of different formats in a unified way.                                                                                         |     Q3 - 2021     |
| UI integration for pluggable API |                                            The UI interacts with a single service (Kubeapps APIs service) to present generic interactions with packages to the user, with format-specific content only where necessary.                                             |     Q3 - 2021     |
|        direct-helm plugin        |                                  plugin-based support for different packaging systems. The direct-Helm plugin aims at just replacing the current Kubeapps logic (mostly implemented by the assetsvc) implemented within a plugin.                                   |     Q3 - 2021     |
|           flux plugin            |                                                                   plugin-based support for different packaging systems, moving across our existing Helm support as well as adding fluxv2 support.                                                                   |     Q3 - 2021     |
|       Update documentation       |                                                                                 Reorganizing and updating Kubeapps docs (including new architecture and plugins); Kubeapps website                                                                                  | Exploring/Ongoing |
|         operators plugin         |                                                                       The operators plugin aims at just replacing the current support for operators in Kubeapps implemented within a plugin.                                                                        | Exploring/Ongoing |
|      kapp-controller plugin      |                                                              plugin-based support for different packaging systems, moving across our existing Helm support as well as adding Carvel packages support.                                                               | Exploring/Ongoing |
|          Improve CI/CD           |                                                                                    Explore a replacement for CircleCI (Github actions or any other alternative); Upgrade tests;                                                                                     | Exploring/Ongoing |
|       Improve auditability       | Design an interface which receives audit events for resources we're interested in and use the built-in support for auditing in Kubernetes (https://kubernetes.io/docs/tasks/debug-application-cluster/audit/) so that the cluster is configured to send the events. | Exploring/Ongoing |
