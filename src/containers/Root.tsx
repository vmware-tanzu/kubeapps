import createHistory from "history/createBrowserHistory";
import * as React from "react";
import { Provider } from "react-redux";
import { Route } from "react-router";
import { ConnectedRouter } from "react-router-redux";

import Dashboard from "../components/Dashboard";
import Layout from "../components/Layout";
import configureStore from "../store";
import ChartList from "./ChartListContainer";
import ChartView from "./ChartViewContainer";

const history = createHistory();
const store = configureStore(history);

class Root extends React.Component {
  public render() {
    return (
      <Provider store={store}>
        <ConnectedRouter history={history}>
          <Layout>
            <section className="routes">
              <Route exact={true} path="/" component={Dashboard} />
              <Route exact={true} path="/charts" component={ChartList} />
              <Route exact={true} path="/charts/:repo" component={ChartList} />
              <Route exact={true} path="/charts/:repo/:id" component={ChartView} />
              <Route
                exact={true}
                path="/charts/:repo/:id/versions/:version"
                component={ChartView}
              />
            </section>
          </Layout>
        </ConnectedRouter>
      </Provider>
    );
  }
}

export default Root;
