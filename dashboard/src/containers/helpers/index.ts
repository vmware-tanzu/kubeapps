import ResourceRef from "../../shared/ResourceRef";
import { IKubeState } from "../../shared/types";

// Takes a set of ResourceRefs and the resources from the Redux state and
// returns the resources related to the ResourceRefs. ResourceRefs that are not
// found are filtered out and the response will only contain resources that are
// available in the state.
export function filterByResourceRefs(refs: ResourceRef[], resources: IKubeState["items"]) {
  return refs.map(r => resources[r.getResourceURL()]).filter(r => r !== undefined);
}
