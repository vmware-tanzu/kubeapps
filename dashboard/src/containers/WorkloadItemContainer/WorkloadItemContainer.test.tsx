import { shallow } from "enzyme";
import * as React from "react";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";

import { IKubeItem, IKubeState, IResource } from "shared/types";
import WorkloadItemContainer from ".";
import WorkloadItem from "../../components/AppView/WorkloadTable/WorkloadTableItem";
import ResourceRef from "../../shared/ResourceRef";

const mockStore = configureMockStore([thunk]);

const makeStore = (resources: { [s: string]: IKubeItem<IResource> }) => {
  const state: IKubeState = {
    items: resources,
    sockets: {},
  };
  return mockStore({ kube: state });
};

describe("WorkloadItemContainer", () => {
  it("maps resource in store to WorkloadItem props", () => {
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
    const wrapper = shallow(
      <WorkloadItemContainer store={store} resourceRef={ref} statusFields={[]} />,
    );
    const form = wrapper.find(WorkloadItem);
    expect(form).toHaveProp({
      name,
      resource: item,
    });
  });
});
