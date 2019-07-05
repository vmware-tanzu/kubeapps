import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import ServiceInstanceView from "../../components/ServiceInstanceView";
import { IServiceInstance } from "../../shared/ServiceInstance";
import { IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      brokerName: string;
      instanceName: string;
      namespace: string;
    };
  };
}

function mapStateToProps({ catalog }: IStoreState, { match: { params } }: IRouteProps) {
  const { instanceName, namespace } = params;
  const { bindingsWithSecrets, instances, classes, plans } = catalog;
  return {
    bindingsWithSecrets,
    errors: catalog.errors,
    instances,
    name: instanceName,
    namespace,
    classes,
    plans,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    addBinding: (bindingName: string, instanceName: string, namespace: string, parameters: {}) =>
      dispatch(actions.catalog.addBinding(bindingName, instanceName, namespace, parameters)),
    deprovision: (instance: IServiceInstance) => dispatch(actions.catalog.deprovision(instance)),
    getPlans: async () => {
      dispatch(actions.catalog.getPlans());
    },
    getClasses: async () => {
      dispatch(actions.catalog.getClasses());
    },
    getInstances: async (ns: string) => {
      dispatch(actions.catalog.getInstances(ns));
    },
    getBindings: async (ns: string) => {
      dispatch(actions.catalog.getBindings(ns));
    },
    removeBinding: (name: string, ns: string) => dispatch(actions.catalog.removeBinding(name, ns)),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ServiceInstanceView);
