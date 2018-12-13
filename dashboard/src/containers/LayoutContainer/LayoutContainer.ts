import { connect } from "react-redux";

import Layout from "../../components/Layout";
import { IStoreState } from "../../shared/types";

function mapStateToProps({ config }: IStoreState) {
  return {
    kubeappsVersion: config.kubeappsVersion,
  };
}

export default connect(mapStateToProps)(Layout);
