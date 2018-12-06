import { shallow } from "enzyme";
import * as React from "react";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { IKubeItem, IKubeState, IResource } from "shared/types";
import SecretsTable from ".";

const mockStore = configureMockStore([thunk]);

const makeStore = (secrets: { [s: string]: IKubeItem<IResource> }) => {
  const state: IKubeState = {
    items: secrets,
  };
  return mockStore({ kube: state });
};

describe("LoginFormContainer props", () => {
  it("maps authentication redux states to props", () => {
    const ns = "wee";
    const name = "foo";
    const item = { isFetching: false, item: { metadata: { name } } as IResource };
    const store = makeStore({
      "v1/secrets/foo": item,
      "v1/pods/foo": item,
    });
    const wrapper = shallow(<SecretsTable store={store} namespace={ns} secretNames={[name]} />);
    const form = wrapper.find("SecretsTable");
    expect(form).toHaveProp({
      namespace: ns,
      secretNames: [name],
      secrets: [item],
    });
  });
});
