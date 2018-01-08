import * as React from 'react';
import { Provider } from 'react-redux';
import { Route } from 'react-router';
import { applyMiddleware, createStore, combineReducers } from 'redux';
import thunkMiddleware from 'redux-thunk';
import createHistory from 'history/createBrowserHistory';
import { ConnectedRouter, routerReducer, routerMiddleware } from 'react-router-redux';

import Layout from '../components/Layout';
import reducer from '../reducers/index';
import { StoreState } from '../store/types';
import Dashboard from '../components/Dashboard';
import Charts from '../containers/ChartsContainer';

const history = createHistory();
const store = createStore<StoreState>(
  combineReducers({
    ...reducer,
    router: routerReducer,
  }),
  applyMiddleware(thunkMiddleware, routerMiddleware(history)));


class Root extends React.Component {
  render() {
    return (
      <Provider store={store}>
        <ConnectedRouter history={history}>
          <Layout>
            <section className="routes">
              <Route exact={true} path="/" component={Dashboard} />
              <Route path="/charts" component={Charts} />
            </section>
          </Layout>
        </ConnectedRouter>
      </Provider>
    );
  }
}

export default Root;
