import { combineReducers } from 'redux';
import { routerReducer } from 'react-router-redux';

import chartsReducer from './charts';
import { StoreState } from '../shared/types';

const rootReducer = combineReducers<StoreState>({
  charts: chartsReducer,
  router: routerReducer,
});

export default rootReducer;
