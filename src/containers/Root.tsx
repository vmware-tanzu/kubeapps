import * as React from 'react';
import { Provider } from 'react-redux';
import { Route } from 'react-router';
import createHistory from 'history/createBrowserHistory';
import { ConnectedRouter } from 'react-router-redux';

import configureStore from '../store';
import Layout from '../components/Layout';
import Dashboard from '../components/Dashboard';
import ChartList from './ChartListContainer';
import ChartView from './ChartViewContainer';

const history = createHistory();
const store = configureStore(history);

class Root extends React.Component {
  render() {
    return (
      <Provider store={store}>
        <ConnectedRouter history={history}>
          <Layout>
            <section className="routes">
              <Route exact={true} path="/" component={Dashboard} />
              <Route exact={true} path="/charts" component={ChartList} />
              <Route exact={true} path="/charts/:repo" component={ChartList} />
              <Route exact={true} path="/charts/:repo/:id" component={ChartView} />
            </section>
          </Layout>
        </ConnectedRouter>
      </Provider>
    );
  }
}

export default Root;
