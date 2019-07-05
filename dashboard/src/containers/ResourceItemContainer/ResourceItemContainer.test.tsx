import { shallow } from "enzyme";
import * as React from "react";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";

import { IKubeItem, IKubeState, IResource } from "shared/types";
import ResourceItemContainer from ".";
import ResourceItem from "../../components/AppView/ResourceTable/ResourceItem/ResourceTableItem";
import ResourceRef from "../../shared/ResourceRef";

const mockStore = configureMockStore([thunk]);

const makeStore = (resources: { [s: string]: IKubeItem<IResource> }) => {
  const state: IKubeState = {
    items: resources,
    sockets: {},
  };
  return mockStore({ kube: state });
};

describe("ResourceItemContainer", () => {
  it("maps resource in store to ResourceItem props", () => {
    const ns = "wee";
    const name = "foo";
    const item = { isFetching: false, item: { metadata: { name } } as IResource };
    const store = makeStore({
      "api/kube/apis/apps/v1/namespaces/wee/statefulsets/foo": item,
    });
    const ref = new ResourceRef({
      apiVersion: "apps/v1",
      kind: "StatefulSet",
      metadata: {
        namespace: ns,
        name,
      },
    } as IResource);
    const wrapper = shallow(<ResourceItemContainer store={store} resourceRef={ref} />);
    const form = wrapper.find(ResourceItem);
    expect(form).toHaveProp({
      name,
      resource: item,
    });
  });
});
