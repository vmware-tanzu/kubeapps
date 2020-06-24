import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";

import Layout from "../../components/Layout";
import { IStoreState } from "../../shared/types";

interface IState extends IStoreState {
  router: RouteComponentProps<{}>;
}

function mapStateToProps({ config: { featureFlags } }: IState) {
  return {
    UI: featureFlags.ui,
  };
}

export default connect(mapStateToProps)(Layout);
