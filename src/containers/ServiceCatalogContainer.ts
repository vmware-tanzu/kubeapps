import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../actions";
import { ServiceCatalogView } from "../components/ServiceCatalog";
import { ServiceCatalog } from "../shared/ServiceCatalog";
import { IStoreState } from "../shared/types";

function mapStateToProps({ catalog }: IStoreState) {
  return {
    ...catalog,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    checkCatalogInstalled: async () => {
      const isInstalled = await ServiceCatalog.isCatalogInstalled();
      isInstalled
        ? dispatch(actions.catalog.installed())
        : dispatch(actions.catalog.notInstalled());
      return isInstalled;
    },
    getCatalog: () => dispatch(actions.catalog.getCatalog()),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(ServiceCatalogView);
