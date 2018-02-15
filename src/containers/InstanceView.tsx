import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../actions";
import { InstanceView } from "../components/InstanceView/InstanceView";
import { IServiceInstance } from "../shared/ServiceInstance";
import { IStoreState } from "../shared/types";

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
  const instance = catalog.instances.find(
    i => i.metadata.name === params.instanceName && i.metadata.namespace === params.namespace,
  );
  const svcClass = instance
    ? catalog.classes.find(c => c.metadata.name === instance.spec.clusterServiceClassRef.name)
    : undefined;
  const svcPlan = instance
    ? catalog.plans.find(p => p.metadata.name === instance.spec.clusterServicePlanRef.name)
    : undefined;

  return {
    bindings: catalog.bindings.filter(
      binding =>
        binding.spec.instanceRef.name === instanceName && binding.metadata.namespace === namespace,
    ),
    instance,
    svcClass,
    svcPlan,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    deprovision: async (instance: IServiceInstance) => {
      await dispatch(actions.catalog.deprovision(instance));
    },
    getCatalog: async () => {
      dispatch(actions.catalog.getCatalog());
    },
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(InstanceView);
