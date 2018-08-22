import { shallow } from "enzyme";
import { Location } from "history";
import * as React from "react";
import { IAuthState } from "reducers/auth";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import LoginForm from "./LoginFormContainer";

const mockStore = configureMockStore([thunk]);

const makeStore = (
  authenticated: boolean,
  authenticating: boolean,
  authenticationError: string,
) => {
  const state: IAuthState = {
    authenticated,
    authenticating,
    authenticationError,
  };
  return mockStore({ auth: state });
};

const emptyLocation: Location = {
  hash: "",
  pathname: "",
  search: "",
  state: "",
};

describe("LoginFormContainer props", () => {
  it("maps authentication redux states to props", () => {
    const store = makeStore(true, true, "It's a trap");
    const wrapper = shallow(<LoginForm store={store} location={emptyLocation} />);
    const form = wrapper.find("LoginForm");
    expect(form).toHaveProp({
      authenticated: true,
      authenticating: true,
      authenticationError: "It's a trap",
    });
  });
});
