// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { get } from "lodash";
import { IResource } from "shared/types";

export const StatefulSetColumns = [
  {
    accessor: "name",
    Header: "NAME",
    getter: (target: IResource) => get(target, "metadata.name"),
  },
  {
    accessor: "desired",
    Header: "DESIRED",
    getter: (target: IResource) => get(target, "spec.replicas"),
  },
  {
    accessor: "upToDate",
    Header: "UP-TO-DATE",
    getter: (target: IResource) => get(target, "status.updatedReplicas"),
  },
  {
    accessor: "ready",
    Header: "READY",
    getter: (target: IResource) => get(target, "status.readyReplicas"),
  },
];
