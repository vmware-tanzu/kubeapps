// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { some } from "lodash";
import { IKubeItem } from "shared/types";

export default function isSomeResourceLoading(resources: Array<IKubeItem<any>>): boolean {
  return some(resources, i => i.isFetching);
}
