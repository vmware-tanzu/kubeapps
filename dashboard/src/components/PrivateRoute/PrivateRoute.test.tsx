import { shallow } from "enzyme";
import { createMemoryHistory } from "history";
import * as React from "react";
import { Redirect, RouteComponentProps } from "react-router-dom";

import PrivateRoute from "./PrivateRoute";

const emptyRouteComponentProps: RouteComponentProps<{}> = {
  history: createMemoryHistory(),
  location: {
    hash: "",
    pathname: "",
    search: "",
    state: "",
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
    push: false,
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
  const wrapper = shallow(
    <PrivateRoute sessionExpired={true} authenticated={false} {...emptyRouteComponentProps} />,
  );
  const renderization: JSX.Element = wrapper.prop("render")();
  expect(renderization.type.toString()).toContain("Modal");
  expect(renderization.props.isOpen).toBe(true);
});
