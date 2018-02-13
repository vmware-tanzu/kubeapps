import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../actions";
import { BrokerView } from "../components/BrokerView";
import { IServiceBroker } from "../shared/ServiceCatalog";
import { IServiceInstance } from "../shared/ServiceInstance";
import { IStoreState } from "../shared/types";

interface IRouteProps {
  match: {
    params: {
      brokerName: string;
    };
  };
}

function mapStateToProps({ catalog }: IStoreState, { match: { params } }: IRouteProps) {
  const broker =
    catalog.brokers.find(
      potental => !!potental.metadata.name.match(new RegExp(params.brokerName, "i")),
    ) || undefined;
  const plans = broker
    ? catalog.plans.filter(
        plan => !!plan.spec.clusterServiceBrokerName.match(new RegExp(broker.metadata.name, "i")),
      )
    : [];
  const classes = broker
    ? catalog.classes.filter(
        serviceClass =>
          !!serviceClass.spec.clusterServiceBrokerName.match(new RegExp(broker.metadata.name, "i")),
      )
    : [];
  const instances = broker ? catalog.instances : [];
  const bindings = broker ? catalog.bindings : [];
  return {
    bindings,
    broker,
    classes,
    instances,
    plans,
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
    sync: async (broker: IServiceBroker) => {
      await dispatch(actions.catalog.sync(broker));
      await dispatch(actions.catalog.getCatalog());
    },
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(BrokerView);
