// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { get } from "lodash";
import { IResource } from "shared/types";

export const DeploymentColumns = [
  {
    accessor: "name",
    Header: "NAME",
    getter: (target: IResource) => get(target, "metadata.name"),
  },
  {
    accessor: "desired",
    Header: "DESIRED",
    getter: (target: IResource) => get(target, "status.replicas"),
  },
  {
    accessor: "upToDate",
    Header: "UP-TO-DATE",
    getter: (target: IResource) => get(target, "status.updatedReplicas"),
  },
  {
    accessor: "available",
    Header: "AVAILABLE",
    getter: (target: IResource) => get(target, "status.availableReplicas"),
  },
];
