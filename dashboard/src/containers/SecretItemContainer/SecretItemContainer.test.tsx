import { shallow } from "enzyme";
import * as React from "react";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";

import { IKubeItem, IKubeState, IResource } from "shared/types";
import SecretItemContainer from ".";
import SecretItem from "../../components/AppView/SecretsTable/SecretItem";
import ResourceRef from "../../shared/ResourceRef";

const mockStore = configureMockStore([thunk]);

const makeStore = (resources: { [s: string]: IKubeItem<IResource> }) => {
  const state: IKubeState = {
    items: resources,
    sockets: {},
  };
  return mockStore({ kube: state });
};

describe("SecretItemContainer", () => {
  it("maps Secret in store to SecretItem props", () => {
    const ns = "wee";
    const name = "foo";
    const item = { isFetching: false, item: { metadata: { name } } as IResource };
    const store = makeStore({
      "api/kube/api/v1/namespaces/wee/secrets/foo": item,
    });
    const ref = new ResourceRef({
      apiVersion: "v1",
      kind: "Secret",
      metadata: {
        namespace: ns,
        name,
      },
    } as IResource);
    const wrapper = shallow(<SecretItemContainer store={store} secretRef={ref} />);
    const form = wrapper.find(SecretItem);
    expect(form).toHaveProp({
      name,
      secret: item,
    });
  });
});
