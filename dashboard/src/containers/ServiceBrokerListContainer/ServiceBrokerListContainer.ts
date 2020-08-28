import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import ServiceBrokerList from "../../components/Config/ServiceBrokerList";
import { IServiceBroker } from "../../shared/ServiceCatalog";
import { IStoreState } from "../../shared/types";

function mapStateToProps({ catalog, config, clusters }: IStoreState) {
  return {
    brokers: catalog.brokers,
    errors: catalog.errors,
    isInstalled: catalog.isServiceCatalogInstalled,
    cluster: clusters.currentCluster,
    kubeappsCluster: config.kubeappsCluster,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    checkCatalogInstalled: async () => {
      dispatch(actions.catalog.checkCatalogInstalled());
    },
    getBrokers: () => dispatch(actions.catalog.getBrokers()),
    sync: (broker: IServiceBroker) => dispatch(actions.catalog.sync(broker)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(ServiceBrokerList);
