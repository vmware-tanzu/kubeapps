import { ConnectedRouter } from "connected-react-router";
import * as React from "react";
import { Provider } from "react-redux";

import store, { history } from "../store";
import ConfigLoaderContainer from "./ConfigLoaderContainer";
import HeaderContainer from "./HeaderContainer";
import Layout from "./LayoutContainer";
import Routes from "./RoutesContainer";

class Root extends React.Component {
  public render() {
    return (
      <Provider store={store}>
        <ConfigLoaderContainer>
          <ConnectedRouter history={history}>
            <Layout headerComponent={HeaderContainer}>
              <Routes />
            </Layout>
          </ConnectedRouter>
        </ConfigLoaderContainer>
      </Provider>
    );
  }
}

export default Root;
