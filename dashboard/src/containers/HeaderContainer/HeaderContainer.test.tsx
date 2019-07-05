import { shallow } from "enzyme";
import * as React from "react";
import { IAuthState } from "reducers/auth";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import Header from "./HeaderContainer";

const mockStore = configureMockStore([thunk]);

const emptyLocation = {
  hash: "",
  pathname: "",
  search: "",
};

const makeStore = (authenticated: boolean, oidcAuthenticated: boolean) => {
  const state: IAuthState = {
    sessionExpired: false,
    authenticated,
    oidcAuthenticated,
    authenticating: false,
  };
  return mockStore({ auth: state, router: { location: emptyLocation } });
};

describe("LoginFormContainer props", () => {
  it("maps authentication redux states to props", () => {
    const store = makeStore(true, true);
    const wrapper = shallow(<Header store={store} />);
    const form = wrapper.find("Header");
    expect(form).toHaveProp({
      authenticated: true,
      hideLogoutLink: true,
    });
  });
});
