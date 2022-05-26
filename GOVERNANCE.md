# Kubeapps Governance

This document defines the project governance for Kubeapps.

## Overview

**Kubeapps**, an open-source project, is committed to building an open, inclusive, productive and self-governing open source community focused on simplifying how applications are deployed and managed in Kubernetes clusters. The community is governed by this document to define how the community should work together to achieve this goal.

## Code of Conduct

The Kubeapps community abides by this [code of conduct](./CODE_OF_CONDUCT.md).

## Community Roles

- **Users**: Members that engage with the Kubeapps community via any medium ([Slack](https://kubernetes.slack.com/messages/kubeapps), [GitHub](https://github.com/vmware-tanzu/kubeapps), etc.).
- **Contributors**: Members contributing to projects (documentation, code reviews, responding to issues, participation in proposal discussions, contributing code, etc.).
- **Maintainers**: Responsible for the overall health and direction of the project; final reviewers of PRs and responsible for releases. Maintainers are expected to contribute code and documentation, ensure the quality of code, triage issues, proactively fix bugs and perform maintenance tasks for Kubeapps components.

## Maintainers

New maintainers must be nominated by an existing maintainer and must be elected by a supermajority of existing maintainers. Likewise, maintainers can be removed by a supermajority of the existing maintainers or can resign by notifying one of the maintainers.

## Decision Making

Ideally, all project decisions are resolved by consensus. If impossible, any maintainer may call a vote. Unless otherwise specified in this document, any vote will be decided by a supermajority of maintainers.

### Supermajority

A supermajority is defined as two-thirds of members in the group. A supermajority of maintainers is required for certain decisions as outlined in this document. A supermajority vote is equivalent to the number of votes in favor being at least twice the number of votes against. A vote to abstain equals not voting at all. For example, if you have 5 maintainers who all cast non-abstaining votes, then a supermajority vote is at least 4 votes in favor. Voting on decisions can happen on the mailing list, GitHub, Slack, email, or via a voting service, when appropriate. Maintainers can either vote "agree, yes, +1", "disagree, no, -1", or "abstain". A vote passes when supermajority is met.

### Lazy Consensus

To maintain velocity in a project as busy as Kubeapps, the concept of [Lazy Consensus](http://en.osswiki.info/concepts/lazy_consensus) is practiced. Ideas and/or proposals should be shared by maintainers via GitHub. Out of respect for other contributors, major changes should also be accompanied by a ping on Slack. Author(s) of the proposal, Pull Requests, issues, etc. will give a time period of no less than five (5) working days for comment and remain cognizant of popular observed world holidays.

Other maintainers may chime in and request additional time for review but should remain cognizant of blocking progress and abstain from delaying progress unless absolutely needed. The expectation is that blocking progress is accompanied by a guarantee to review and respond to the relevant action(s) (proposals, PRs, issues, etc.) in short order.

Lazy consensus does not apply to the process of:

- Removal of maintainers from Kubeapps

## Proposal Process

The proposal process is defined [here](./site/content/docs/latest/reference/proposals/proposals.md).

## Updating Governance

All substantive changes in Governance require a supermajority agreement by all [maintainers](./MAINTAINERS.md).

## Credits

Sections of this document have been borrowed from [Velero](https://github.com/vmware-tanzu/velero) and [Pinniped](https://github.com/vmware-tanzu/pinniped) projects.
