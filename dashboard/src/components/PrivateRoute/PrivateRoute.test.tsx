// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsModal } from "@cds/react/modal";
import { shallow } from "enzyme";
import { createMemoryHistory } from "history";
import React from "react";
import { Redirect, RouteComponentProps } from "react-router-dom";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import PrivateRoute from "./PrivateRoute";

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

class MockComponent extends React.Component {}

it("redirects to the /login route if not authenticated", () => {
  const wrapper = shallow(
    <PrivateRoute
      sessionExpired={false}
      authenticated={false}
      path="/test"
      component={MockComponent}
      {...emptyRouteComponentProps}
    />,
  );
  const RenderMethod = (wrapper.instance() as PrivateRoute).renderRouteIfAuthenticated;
  const wrapper2 = shallow(<RenderMethod {...emptyRouteComponentProps} />);
  expect(wrapper2.find(Redirect).exists()).toBe(true);
  expect(wrapper2.find(Redirect).props()).toMatchObject({
    to: { pathname: "/login" },
  } as any);
});

it("renders the given component when authenticated", () => {
  const wrapper = shallow(
    <PrivateRoute
      sessionExpired={false}
      authenticated={true}
      path="/test"
      component={MockComponent}
      {...emptyRouteComponentProps}
    />,
  );
  const RenderMethod = (wrapper.instance() as PrivateRoute).renderRouteIfAuthenticated;
  const wrapper2 = shallow(<RenderMethod {...emptyRouteComponentProps} />);
  expect(wrapper2.find(MockComponent).exists()).toBe(true);
});

it("renders modal to reload the page if the session is expired", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <PrivateRoute sessionExpired={true} authenticated={false} {...emptyRouteComponentProps} />,
  );
  expect(wrapper.find(CdsModal)).toExist();
});

it("does not render modal to reload the page if the session isn't expired", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <PrivateRoute sessionExpired={false} authenticated={false} {...emptyRouteComponentProps} />,
  );
  expect(wrapper.find(CdsModal)).not.toExist();
});
