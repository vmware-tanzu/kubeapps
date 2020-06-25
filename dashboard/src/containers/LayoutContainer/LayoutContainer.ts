import { connect } from "react-redux";

import Layout from "../../components/Layout";
import { IStoreState } from "../../shared/types";

function mapStateToProps({ config: { featureFlags } }: IStoreState) {
  return {
    UI: featureFlags.ui,
  };
}

export default connect(mapStateToProps)(Layout);
