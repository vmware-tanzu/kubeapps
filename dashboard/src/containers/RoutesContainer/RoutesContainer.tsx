import { connect } from "react-redux";
import { withRouter } from "react-router";

import Routes from "../../components/Routes";
import { IStoreState } from "../../shared/types";

function mapStateToProps({ namespace }: IStoreState) {
  return { namespace: namespace.current };
}

export default withRouter(connect(mapStateToProps)(Routes));
