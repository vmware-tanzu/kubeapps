// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { get } from "lodash";
import { IResource } from "shared/types";

export const OtherResourceColumns = [
  {
    accessor: "name",
    Header: "NAME",
    getter: (target: IResource) => get(target, "metadata.name"),
  },
  {
    accessor: "kind",
    Header: "KIND",
    getter: (target: IResource) => get(target, "kind"),
  },
];
