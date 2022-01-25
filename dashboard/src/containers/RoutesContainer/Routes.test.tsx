// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { deepClone } from "@cds/core/internal/utils/identity";
import AlertGroup from "components/AlertGroup";
import LoadingWrapper from "components/LoadingWrapper";
import { mount } from "enzyme";
import { createMemoryHistory } from "history";
import { Provider } from "react-redux";
import { StaticRouter, Redirect, RouteComponentProps } from "react-router-dom";
import { IFeatureFlags } from "shared/Config";
import { defaultStore } from "shared/specs/mountWrapper";
import { app } from "shared/url";
import NotFound from "../../components/NotFound";
import Routes from "./Routes";

const emptyRouteComponentProps: RouteComponentProps<{}> = {
  history: createMemoryHistory(),
  location: {
    hash: "",
    pathname: "",
    search: "",
    state: "",
    key: "",
  },
  match: {
    isExact: false,
    params: {},
    path: "",
    url: "",
  },
};

const defaultFeatureFlags: IFeatureFlags = {
  operators: false,
};

it("invalid path should show a 404 error", () => {
  const wrapper = mount(
    <StaticRouter location="/random" context={{}}>
      <Routes
        {...emptyRouteComponentProps}
        cluster={"default"}
        currentNamespace={"default"}
        authenticated={true}
        featureFlags={defaultFeatureFlags}
      />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).toExist();
  expect(wrapper.text()).toContain("The page you are looking for can't be found.");
});

it("should render a redirect to the default cluster and namespace", () => {
  const wrapper = mount(
    <StaticRouter location="/" context={{}}>
      <Routes
        {...emptyRouteComponentProps}
        cluster={"default"}
        currentNamespace={"default"}
        authenticated={true}
        featureFlags={defaultFeatureFlags}
      />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).not.toExist();
  expect(wrapper.find(Redirect).prop("to")).toEqual({
    pathname: app.apps.list("default", "default"),
  });
});

it("should render a redirect to the login page", () => {
  const wrapper = mount(
    <StaticRouter location="/" context={{}}>
      <Routes
        {...emptyRouteComponentProps}
        cluster={""}
        currentNamespace={""}
        authenticated={false}
        featureFlags={defaultFeatureFlags}
      />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).not.toExist();
  expect(wrapper.find(Redirect).prop("to")).toEqual({ pathname: "/login" });
});

it("should render a redirect to the login page (even with cluster or ns info)", () => {
  const wrapper = mount(
    <StaticRouter location="/" context={{}}>
      <Routes
        {...emptyRouteComponentProps}
        cluster={"default"}
        currentNamespace={"default"}
        authenticated={false}
        featureFlags={defaultFeatureFlags}
      />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).not.toExist();
  expect(wrapper.find(Redirect).prop("to")).toEqual({ pathname: "/login" });
});

it("should render a loading wrapper if authenticated but the cluster and ns info is not populated", () => {
  const wrapper = mount(
    <StaticRouter location="/" context={{}}>
      <Routes
        {...emptyRouteComponentProps}
        cluster={""}
        currentNamespace={""}
        authenticated={true}
        featureFlags={defaultFeatureFlags}
      />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).not.toExist();
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("should render a warning message if operators are deactivated", () => {
  const componentProps = deepClone(emptyRouteComponentProps);
  componentProps.featureFlags = { operators: false };
  const operatorsUrl = app.config.operators("default", "default");

  const wrapper = mount(
    <StaticRouter location={operatorsUrl} context={{}}>
      <Routes
        {...componentProps}
        cluster={""}
        currentNamespace={""}
        authenticated={true}
        featureFlags={defaultFeatureFlags}
      />
    </StaticRouter>,
  );
  expect(wrapper.find(AlertGroup)).toExist();
  expect(wrapper.find(AlertGroup).text()).toBe(
    "Operators support has been deactivated by default for Kubeapps. It can be enabled in values configuration.",
  );
});

it("should route to operators if enabled", () => {
  const componentProps = deepClone(emptyRouteComponentProps);
  componentProps.featureFlags = { operators: true };
  const operatorsUrl = app.config.operators("default", "default");
  const wrapper = mount(
    <Provider store={defaultStore}>
      <StaticRouter location={operatorsUrl} context={{}}>
        <Routes
          {...componentProps}
          cluster={""}
          currentNamespace={""}
          authenticated={true}
          featureFlags={{ operators: true }}
        />
      </StaticRouter>
    </Provider>,
  );
  expect(wrapper.find(AlertGroup)).not.toExist();
});
