import * as _ from "lodash";

import { IKubeItem } from "./../../../shared/types";

export default function isSomeResourceLoading(resources: Array<IKubeItem<any>>): boolean {
  return _.some(resources, i => i.isFetching);
}
