import * as _ from "lodash";

import ResourceRef from "../../shared/ResourceRef";
import { IKubeState } from "../../shared/types";

export function filterByResourceRefs(refs: ResourceRef[], resources: IKubeState["items"]) {
  return refs.map(r => resources[r.getResourceURL()]).filter(r => r !== undefined);
}
