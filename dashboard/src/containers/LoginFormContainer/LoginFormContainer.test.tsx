// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import { Location } from "history";
import { IAuthState } from "reducers/auth";
import { IClustersState } from "reducers/cluster";
import { IConfigState } from "reducers/config";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { SupportedThemes } from "shared/Config";
import { IStoreState } from "shared/types";
import LoginForm from "./LoginFormContainer";

const mockStore = configureMockStore([thunk]);

const makeStore = (
  sessionExpired: boolean,
  authenticated: boolean,
  authenticating: boolean,
  oidcAuthenticated: boolean,
  authenticationError: string,
  authProxyEnabled: boolean,
  oauthLoginURI: string,
) => {
  const auth: IAuthState = {
    sessionExpired,
    authenticated,
    authenticating,
    oidcAuthenticated,
    authenticationError,
  };
  const config: IConfigState = {
    authProxyEnabled,
    oauthLoginURI,
    loaded: true,
    kubeappsCluster: "",
    kubeappsNamespace: "",
    helmGlobalNamespace: "",
    carvelGlobalNamespace: "",
    appVersion: "",
    oauthLogoutURI: "",
    featureFlags: { operators: false },
    clusters: [],
    authProxySkipLoginPage: false,
    theme: SupportedThemes.light,
    remoteComponentsUrl: "",
    customAppViews: [],
    skipAvailablePackageDetails: false,
    createNamespaceLabels: {},
    configuredPlugins: [],
  };
  const clusters: IClustersState = {
    currentCluster: "default",
    clusters: {
      default: {
        currentNamespace: "default",
        namespaces: [],
        canCreateNS: true,
      },
    },
  };
  return mockStore({ auth, config, clusters } as Partial<IStoreState>);
};

const emptyLocation: Location = {
  hash: "",
  pathname: "",
  search: "",
  state: "",
  key: "",
};

describe("LoginFormContainer props", () => {
  it("maps authentication redux states to props", () => {
    const authProxyEnabled = true;
    const store = makeStore(
      true,
      true,
      true,
      true,
      "It's a trap",
      authProxyEnabled,
      "/myoauth/start",
    );
    const wrapper = shallow(<LoginForm store={store} location={emptyLocation} />);
    const form = wrapper.find("LoginForm");
    expect(form).toHaveProp({
      authenticated: true,
      authenticating: true,
      authenticationError: "It's a trap",
      oauthLoginURI: "/myoauth/start",
    });
  });

  it("does not receive oauthLoginURI if authProxyEnabled is false", () => {
    const authProxyEnabled = false;
    const store = makeStore(
      true,
      true,
      true,
      true,
      "It's a trap",
      authProxyEnabled,
      "/myoauth/start",
    );
    const wrapper = shallow(<LoginForm store={store} location={emptyLocation} />);
    const form = wrapper.find("LoginForm");
    expect(form).toHaveProp({
      authenticated: true,
      authenticating: true,
      authenticationError: "It's a trap",
      oauthLoginURI: "",
    });
  });
});
