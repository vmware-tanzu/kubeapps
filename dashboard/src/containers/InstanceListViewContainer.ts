import * as qs from "qs";
import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { Dispatch } from "redux";

import actions from "../actions";
import { InstanceListView } from "../components/InstanceListView";
import { IStoreState } from "../shared/types";

function mapStateToProps(
  { catalog, namespace }: IStoreState,
  { location }: RouteComponentProps<{ brokerName: string }>,
) {
  const brokers = catalog.brokers;
  const plans = catalog.plans;
  const classes = catalog.classes;
  const instances = catalog.instances;
  const isInstalled = catalog.isInstalled;
  const showAlphaWarning = catalog.showAlphaWarning;
  return {
    brokers,
    classes,
    error: catalog.errors.fetch,
    filter: qs.parse(location.search, { ignoreQueryPrefix: true }).q || "",
    instances,
    isInstalled,
    namespace: namespace.current,
    plans,
    showAlphaWarning,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    checkCatalogInstalled: async () => {
      dispatch(actions.catalog.checkCatalogInstalled());
    },
    disableAlphaWarning: () => dispatch(actions.catalog.disableAlphaWarning()),
    getCatalog: async (ns: string) => {
      dispatch(actions.catalog.getCatalog(ns));
    },
    pushSearchFilter: (filter: string) => dispatch(actions.shared.pushSearchFilter(filter)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(InstanceListView);
