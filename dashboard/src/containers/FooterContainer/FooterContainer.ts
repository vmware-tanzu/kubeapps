import { connect } from "react-redux";

import Footer from "../../components/Footer";
import { IStoreState } from "../../shared/types";

function mapStateToProps({ config }: IStoreState) {
  return {
    appVersion: config.appVersion,
  };
}

export default connect(mapStateToProps)(Footer);
