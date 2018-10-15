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
  const instance = catalog.instances.list.find(
    i => i.metadata.name === params.instanceName && i.metadata.namespace === params.namespace,
  );
  const svcClass = instance
    ? catalog.classes.list.find(
        c =>
          !!instance.spec.clusterServiceClassRef &&
          c.metadata.name === instance.spec.clusterServiceClassRef.name,
      )
    : undefined;
  const svcPlan = instance
    ? catalog.plans.find(
        p =>
          !!instance.spec.clusterServicePlanRef &&
          p.metadata.name === instance.spec.clusterServicePlanRef.name,
      )
    : undefined;

  return {
    bindingsWithSecrets: catalog.bindingsWithSecrets.filter(
      b =>
        b.binding.spec.instanceRef.name === instanceName &&
        b.binding.metadata.namespace === namespace,
    ),
    errors: catalog.errors,
    instance,
    name: instanceName,
    namespace,
    svcClass,
    svcPlan,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    addBinding: (bindingName: string, instanceName: string, namespace: string, parameters: {}) =>
      dispatch(actions.catalog.addBinding(bindingName, instanceName, namespace, parameters)),
    deprovision: (instance: IServiceInstance) => dispatch(actions.catalog.deprovision(instance)),
    getCatalog: (ns: string) => {
      dispatch(actions.catalog.getCatalog(ns));
    },
    removeBinding: (name: string, ns: string) => dispatch(actions.catalog.removeBinding(name, ns)),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ServiceInstanceView);
