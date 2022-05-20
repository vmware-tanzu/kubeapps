## **Kubeapps Project Roadmap**

###

**About this document**

This document provides a link to the[ Kubeapps Project issues](https://github.com/vmware-tanzu/kubeapps/issues) list that serves as the up to date description of items that are in the Kubeapps release pipeline. This should serve as a reference point for Kubeapps users and contributors to understand where the project is heading, and help determine if a contribution could be conflicting with a longer term plan.

###

**How to help?**

Discussion on the roadmap can take place in threads under [Issues](https://github.com/vmware-tanzu/kubeapps/issues). Please open and comment on an issue if you want to provide suggestions and feedback to an item in the roadmap. Please review the roadmap to avoid potential duplicated effort.

###

**How to add an item to the roadmap?**

Please open an issue to track any initiative on the roadmap of Kubeapps (usually driven by new feature requests). We will work with and rely on our community to focus our efforts to improve Kubeapps.

###

**Current Roadmap**

The following table includes the current roadmap for Kubeapps. If you have any questions or would like to contribute to Kubeapps, please contact us by Slack on the [#Kubeapps channel](https://kubernetes.slack.com/messages/kubeapps) to discuss with our team.

If you don't know where to start, we are always looking for contributors that will help us reduce technical, automation, and documentation debt. Please take the timelines & dates as proposals and goals. Priorities and requirements change based on community feedback, roadblocks encountered, community contributions, etc. If you depend on a specific item, we encourage you to contact us by Slack, or help us deliver that feature by contributing to Kubeapps.

Last Updated: May 2022
Epic|Description|Timeline|
|--|--|--|
| [Distribute Kubeapps Carvel package as part of TCE](https://github.com/vmware-tanzu/kubeapps/milestone/40) | Include the Kubeapps application in the TCE repository so that users can view and interact with the packages available in the TCE repository via the Kubeapps UI, which now supports browsing and installing Carvel packages | FY23-Q2 |
| [New Kubeapps website](https://github.com/vmware-tanzu/kubeapps/milestone/37) | Build a new Kubeapps website aligned with the rest of open-sourced projects in Tanzu | FY23-Q2 |
| [Implementation of package repository for Helm plugin](https://github.com/vmware-tanzu/kubeapps/milestone/42) | Once package repository API has been defined, it is time to implement it for Helm plugin | FY23-Q2 |
| [Implementation of package repository for Carvel plugin](https://github.com/vmware-tanzu/kubeapps/milestone/43) | Once package repository API has been defined, it is time to implement it for Carvel plugin | FY23-Q2 |
| [Design how to enable plugins to customize UI](https://github.com/vmware-tanzu/kubeapps/milestone/46) | Kubeapps UI should allow to include some features provided by specific plugins in addition to the core features provided by all of the packaging plugins | FY23-Q2 |
| [Standardize caching repo and chart data in Kubeapps](https://github.com/vmware-tanzu/kubeapps/milestone/45) | Look into replacing postgresql as a mechanism for helm plug-in with redis to be consistent with flux support in kubeapps (which already uses redis as a caching mechanism) | FY23-Q2 |
| Kubeapps Docker Desktop extension | Develop and publish a Docker Desktop extension to simplify the onboarding to Kubeapps for Docker Desktop users | FY-23 |
