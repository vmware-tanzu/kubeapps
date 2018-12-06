import * as _ from "lodash";

import { IKubeItem } from "./../../../shared/types";

export default function isSomeResourceLoading(resources: Array<IKubeItem<any>>): boolean {
  let isFetching = false;
  _.each(resources, i => {
    if (i.isFetching) {
      isFetching = true;
    }
  });
  return isFetching;
}
