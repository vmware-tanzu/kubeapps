import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../actions";
import { ServiceCatalogView } from "../components/Config/ServiceCatalog";
import { IServiceBroker } from "../shared/ServiceCatalog";
import { IStoreState } from "../shared/types";

function mapStateToProps({ catalog }: IStoreState) {
  return {
    ...catalog,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    checkCatalogInstalled: async () => {
      dispatch(actions.catalog.checkCatalogInstalled());
    },
    getCatalog: () => dispatch(actions.catalog.getCatalog()),
    sync: (broker: IServiceBroker) => dispatch(actions.catalog.sync(broker)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(ServiceCatalogView);
