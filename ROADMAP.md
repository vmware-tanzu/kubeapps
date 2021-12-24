## **Kubeapps Project Roadmap**

###

**About this document**

This document provides a link to the[ Kubeapps Project issues](https://github.com/kubeapps/kubeapps/issues) list that serves as the up to date description of items that are in the Kubeapps release pipeline. This should serve as a reference point for Kubeapps users and contributors to understand where the project is heading, and help determine if a contribution could be conflicting with a longer term plan.

###

**How to help?**

Discussion on the roadmap can take place in threads under [Issues](https://github.com/kubeapps/kubeapps/issues). Please open and comment on an issue if you want to provide suggestions and feedback to an item in the roadmap. Please review the roadmap to avoid potential duplicated effort.

###

**How to add an item to the roadmap?**

Please open an issue to track any initiative on the roadmap of Kubeapps (usually driven by new feature requests). We will work with and rely on our community to focus our efforts to improve Kubeapps.

###

**Current Roadmap**

The following table includes the current roadmap for Kubeapps. If you have any questions or would like to contribute to Kubeapps, please contact by Slack on the [#Kubeapps channel](https://kubernetes.slack.com/messages/kubeapps) to discuss with our team. If you don't know where to start, we are always looking for contributors that will help us reduce technical, automation, and documentation debt. Please take the timelines & dates as proposals and goals. Priorities and requirements change based on community feedback, roadblocks encountered, community contributions, etc. If you depend on a specific item, we encourage you to contact by Slack, or help us deliver that feature by contributing to Kubeapps.

Last Updated: December 2021
Theme|Description|Timeline|
|--|--|--|
|Kubernetes API Service |Currently when an app is installed on the cluster, our AppView gets the data about the related k8s resources from the k8s api server. We want to replace that with an API call which allows getting the (relevant) resources for a particular installed package (only). |Q4 - 2021|
|flux plugin |Plugin-based support for different packaging systems, moving across our existing Helm support as well as adding fluxv2 support. |Q4 - 2021|
|kapp-controller plugin |Plugin-based support for different packaging systems, moving across our existing Helm support as well as adding Carvel packages support. |Q4 - 2021|
|Package repository API |Define a package repositories API with similar core interface to packages API. |Q4 - 2021|
|Improve CI/CD and Release process | Explore a replacement for CircleCI (Github actions or any other alternative); Upgrade tests; |Q4 -2021|
|Update documentation |Reorganizing and updating Kubeapps docs (including new architecture and plugins); New Kubeapps website |Ongoing|
|operators plugin |The operators plugin aims at just replacing the current support for operators in Kubeapps implemented within a plugin. |Backlog|
|Improve auditability |Design an interface which receives audit events for resources we're interested in and use the built-in support for auditing in Kubernetes (https://kubernetes.io/docs/tasks/debug-application-cluster/audit/) so that the cluster is configured to send the events. |Backlog|
