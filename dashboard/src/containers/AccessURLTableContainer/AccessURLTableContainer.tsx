import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import ResourceRef from "shared/ResourceRef";
import { IStoreState } from "shared/types";
import actions from "../../actions";
import AccessURLTable from "../../components/AppView/AccessURLTable";
import { filterByResourceRefs } from "../helpers";

interface IAccessURLTableContainerProps {
  serviceRefs: ResourceRef[];
  ingressRefs: ResourceRef[];
}

function mapStateToProps({ kube }: IStoreState, props: IAccessURLTableContainerProps) {
  // Extract the Services and Ingresses form the Redux state using the keys for
  // each ResourceRef.
  return {
    services: filterByResourceRefs(props.serviceRefs, kube.items),
    ingresses: filterByResourceRefs(props.ingressRefs, kube.items),
    ingressRefs: props.ingressRefs,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    getResource: (r: ResourceRef) => dispatch(actions.kube.getResource(r)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AccessURLTable);
