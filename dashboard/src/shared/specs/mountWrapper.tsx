import { RouterState } from "connected-react-router";
import { mount } from "enzyme";
import { cloneDeep, merge } from "lodash";
import { IntlProvider } from "react-intl";
import { Provider } from "react-redux";
import { BrowserRouter as Router } from "react-router-dom";
import { initialState as appsInitialState } from "reducers/apps";
import { initialState as authInitialState } from "reducers/auth";
import { initialState as packagesInitialState } from "reducers/packages";
import { initialState as clustersInitialState } from "reducers/cluster";
import { initialState as configInitialState } from "reducers/config";
import { initialState as kubeInitialState } from "reducers/kube";
import { operatorsInitialState } from "reducers/operators";
import { initialState as reposInitialState } from "reducers/repos";
import configureMockStore, { MockStore } from "redux-mock-store";
import thunk from "redux-thunk";
import I18n from "shared/I18n";
import { IStoreState } from "shared/types";

const mockStore = configureMockStore([thunk]);

export const initialState = {
  apps: cloneDeep(appsInitialState),
  auth: cloneDeep(authInitialState),
  packages: cloneDeep(packagesInitialState),
  config: {
    ...cloneDeep(configInitialState),
    kubeappsCluster: "default-cluster",
    kubeappsNamespace: "kubeapps",
  },
  kube: cloneDeep(kubeInitialState),
  clusters: {
    ...cloneDeep(clustersInitialState),
    currentCluster: "default-cluster",
    clusters: {
      "default-cluster": {
        currentNamespace: "default",
        namespaces: ["default", "other"],
        canCreateNS: true,
      },
      "second-cluster": {
        currentNamespace: "default",
        namespaces: ["default", "other"],
        canCreateNS: true,
      },
    },
  },
  repos: cloneDeep(reposInitialState),
  operators: cloneDeep(operatorsInitialState),
  router: {} as RouterState,
} as IStoreState;

export const defaultStore = mockStore(initialState);

// Default to english 1i8nconfiguration
const messages = I18n.getDefaultConfig().messages;
const locale = I18n.getDefaultConfig().locale;

// getStore returns a store initialised with a merge of
// the initial state with any passed extra state.
export const getStore = (extraState: object) => {
  const state = cloneDeep(initialState);
  return mockStore(merge(state, extraState));
};

export const mountWrapper = (store: MockStore, children: React.ReactElement) =>
  mount(
    <Provider store={store}>
      <IntlProvider locale={locale} key={locale} messages={messages} defaultLocale={locale}>
        <Router>{children}</Router>
      </IntlProvider>
      ,
    </Provider>,
  );
