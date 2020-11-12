import Header from "components/Header";
import Layout from "components/Layout";
import { ConnectedRouter } from "connected-react-router";
import * as React from "react";
import { Provider } from "react-redux";

import store, { history } from "../store";
// TODO(andresmgot): Containers should be no longer needed, replace them when possible
import ConfigLoaderContainer from "./ConfigLoaderContainer";
import Routes from "./RoutesContainer";

class Root extends React.Component {
  public render() {
    return (
      <Provider store={store}>
        <ConfigLoaderContainer>
          <ConnectedRouter history={history}>
            <React.Suspense fallback={null}>
              <Layout headerComponent={Header}>
                <Routes />
              </Layout>
            </React.Suspense>
          </ConnectedRouter>
        </ConfigLoaderContainer>
      </Provider>
    );
  }
}

export default Root;
