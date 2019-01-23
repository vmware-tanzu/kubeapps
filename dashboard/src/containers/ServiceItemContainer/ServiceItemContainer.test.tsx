import { shallow } from "enzyme";
import * as React from "react";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";

import { IKubeItem, IKubeState, IResource } from "shared/types";
import ServiceItemContainer from ".";
import ServiceItem from "../../components/AppView/ServicesTable/ServiceItem";
import ResourceRef from "../../shared/ResourceRef";

const mockStore = configureMockStore([thunk]);

const makeStore = (secrets: { [s: string]: IKubeItem<IResource> }) => {
  const state: IKubeState = {
    items: secrets,
  };
  return mockStore({ kube: state });
};

describe("ServiceItemContainer", () => {
  it("maps Service in store to ServiceItem props", () => {
    const ns = "wee";
    const name = "foo";
    const item = { isFetching: false, item: { metadata: { name } } as IResource };
    const store = makeStore({
      "api/kube/api/v1/namespaces/wee/services/foo": item,
    });
    const ref = new ResourceRef({
      apiVersion: "v1",
      kind: "Service",
      metadata: {
        namespace: ns,
        name,
      },
    } as IResource);
    const wrapper = shallow(<ServiceItemContainer store={store} serviceRef={ref} />);
    const form = wrapper.find(ServiceItem);
    expect(form).toHaveProp({
      name,
      service: item,
    });
  });
});
