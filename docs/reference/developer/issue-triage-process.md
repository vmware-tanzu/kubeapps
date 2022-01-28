# Kubeapps triage process

## Kubeapps backlog

Kubeapps keeps a backlog of issues on GitHub submitted both by maintainers and contributors: this backlog comprises bugs, feature requests, and technical debt. If it’s a bug, an idea for a new feature, or something in between, it’s filed as an issue in the [kubeapps issue page](https://github.com/kubeapps/kubeapps/issues) on GitHub.
There are some special considerations about how the Kubeapps maintainer team manages its backlog on GitHub:

- The issue repository is completely open. The maintainer team, along with the entire Kubeapps community, files all feature enhancements, bugs, and potential future work into the open repository.
- Issues are closed if solved, or outdated (meaning the issue does not apply anymore according to the evolution of Kubeapps or it was waiting for more information which was never received).

## Kubeapps' triage process

The [kubeapps/kubeapps](https://github.com/kubeapps/kubeapps) issues are triaged by any member of the maintainer team. The triage process is defined with the objective of helping the Kubeapps team during the planning of each iteration so that the selected issues provide the highest added value to the product itself according to the Kubeapps maintainer team capacities.

The triage process should be performed manually for any issue in the Inbox column and will consist of:

- Reading the description of the issue.
  - If more information is requested, as a result of the triage process the issue must be labeled as '<awaiting-more-evidence>', a comment requesting for information added and the issue must remain in the Inbox column.
- Adding labels for each category (required): component, kind, size and priority.
- Check if it is a 'good-first-issue' to start contributing to Kubeapps and label the issue as such.
- Moving the issue to the appropriate column according to the triage process:
  - **Committed** → Issues labeled as 'priority/unbreak-now'
  - **Next** iteration discussion → Issues labeled as 'priority/high'
  - **Inbox** → Issues labeled 'priority/medium' and 'priority/low' will remain in the inbox before moving to the **Backlog** to double-check with the whole maintainer team during the iteration planning meeting.
  - **Inbox** → Issues labeled as 'awaiting-more-evidence'
- Repeat until all issues in the **Inbox** column are triaged.

Besides this manual triage process, there are two automatic process lead by the bots configured in GitHub:

- Automatically move new and reopened issues to the **Inbox** column to start the triage process (due to eventual project-bot outages, once a month we should manually search for issues with "no:project" assigned).
- Automatically **dependabot** and **Snyk** are labelling PRs depending on the target language.
- Automatically **stalebot** is checking inactive issues to label them as 'stale'. An issue becomes stale after 15 days of inactivity and the bot will close it after 3 days of inactivity for stale issues. To be considered:
  - Issues labeled as 'priority/unbreak-now' → 'priority/high' → 'priority/medium' will never be labeled as 'stale'.
  - Issues labeled as 'kind/feature', 'kind/bug' or 'kind/refactor' will never be labeled as 'stale'
  - Only issues labeled as 'priority/low' or 'awaiting-more-evidence' could be considered stale (and those without any priority label).
  - The label to use when marking an issue as stale is 'stale'.

## Labels

### Component

| component/                 |                |            |           |           |                  |
| -------------------------- | -------------- | ---------- | --------- | --------- | ---------------- |
| 'apprepository-controller' | 'asset-syncer' | 'assetsvc' | 'auth'    | 'chart'   | 'ci'             |
| 'dashboard'                | 'docs'         | 'hub'      | 'kubeops' | 'project' | 'pinniped-proxy' |

### Kind

| kind/ |           |            |            |
| ----- | --------- | ---------- | ---------- |
| 'bug' | 'feature' | 'question' | 'refactor' |

### Size

| size/ |                                                                                             |
| ----- | ------------------------------------------------------------------------------------------- |
| 'XS'  | A task that can be done by a person in less than 1 full day                                 |
| 'S'   | A story that can be done by a person in 1-3 days, with no uncertainty                       |
| 'M'   | A story that can be done by a person in 4-7 days, possibly with some uncertainty            |
| 'L'   | A story that requires investigation and possibly will take a person a full 2-week iteration |
| 'XL'  | A story too big or with too many unknowns. Needs investigation and split into several ones  |

### Priority

| priority/     |        |          |       |
| ------------- | ------ | -------- | ----- |
| 'unbreak-now' | 'high' | 'medium' | 'low' |

**Contribution labels**:

'awaiting-more-evidence' - information requested to the reporter.

'help-wanted' - the maintainer team wants help on an issue or pull request.

'good-first-issue' - good first issues to start contributing to Kubeapps.

| Automatic labels |      |              |        |            |         |
| ---------------- | ---- | ------------ | ------ | ---------- | ------- |
| 'dependencies'   | 'go' | 'javascript' | 'rust' | 'security' | 'stale' |

## Kubeapps milestones

Next step in the triage process to help the community understand the project roadmap is to define milestones and associate issues to them.
According to the Kubeapps team practices, milestones will be defined based on quarters and years according to the following pattern: **YYYY-QN** (i.e. 2021-Q1 - 2021-Q2 - etc), setting the due date for each milestone.

## Kubeapps iteration planning guidelines

1. Review issues in column **Inbox** (untriaged issues, awaiting-more-evidence, triaged issues labeled as low and medium priority) and move them to the column according to the triage process. Issues labeled as 'awaiting-more-evidence' must be checked if updated to be triaged.
2. Review issues in column **Next iteration discussion** and decide what issues should be moved to the **Committed** column according to the capacity and uncompleted issues from previous iterations (**In progress**).
3. Filter issues by 'priority/unbreak-now' → Check that all issues labeled as 'priority/unbreak-now' are, at least, placed in the **Committed** for next iteration column.
4. Filter issues by 'priority/high' → Check that all issues labeled as 'priority/high' are, at least, placed in the “**Next iteration discussion**” column. If any of the 'priority/high' issues shouldn’t be discussed for the next iteration it means that they should be re-prioritized and moved back to the **Backlog**.
5. Filter issues by 'priority/medium' → Check if any of the issues should be re-prioritize or/and added to the **Next iteration discussion** (or **Committed**).
6. Filter issues by 'priority/low' → Check if any of the issues should be re-prioritize or/and added to the **Next iteration discussion** (or **Committed**).
