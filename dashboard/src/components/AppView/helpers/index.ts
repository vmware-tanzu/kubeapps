import { some } from "lodash";

import { IKubeItem } from "./../../../shared/types";

export default function isSomeResourceLoading(resources: Array<IKubeItem<any>>): boolean {
  return some(resources, i => i.isFetching);
}
