import { applyMiddleware, createStore } from 'redux';
import { routerMiddleware } from 'react-router-redux';
import thunkMiddleware from 'redux-thunk';
import { History } from 'history';

import rootReducer from '../reducers';
import { StoreState } from '../shared/types';

const configureStore = (history: History) => createStore<StoreState>(
  rootReducer,
  applyMiddleware(thunkMiddleware, routerMiddleware(history)),
);

export default configureStore;
