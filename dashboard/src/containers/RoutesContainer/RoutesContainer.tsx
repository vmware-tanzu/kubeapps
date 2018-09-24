import { connect } from "react-redux";
import { withRouter } from "react-router";

import { IStoreState } from "../../shared/types";
import Routes from "./Routes";

function mapStateToProps({ namespace }: IStoreState) {
  return { namespace: namespace.current };
}

export default withRouter(connect(mapStateToProps)(Routes));
