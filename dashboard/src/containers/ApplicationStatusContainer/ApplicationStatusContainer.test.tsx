import { shallow } from "enzyme";
import * as React from "react";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";

import ApplicationStatusContainer from ".";
import ApplicationStatus from "../../components/ApplicationStatus";
import ResourceRef from "../../shared/ResourceRef";
import { IKubeItem, IKubeState, IResource } from "../../shared/types";

const mockStore = configureMockStore([thunk]);

const makeStore = (resources: { [s: string]: IKubeItem<IResource> }) => {
  const state: IKubeState = {
    items: resources,
    sockets: {},
  };
  return mockStore({ kube: state });
};

describe("ApplicationStatusContainer", () => {
  it("maps Deployment in store to ApplicationStatus props", () => {
    const ns = "wee";
    const name = "foo";
    const item = { isFetching: false, item: { metadata: { name } } as IResource };
    const store = makeStore({
      "api/kube/apis/apps/v1/namespaces/wee/deployments/foo": item,
    });
    const ref = new ResourceRef({
      apiVersion: "apps/v1",
      kind: "Deployment",
      metadata: {
        namespace: ns,
        name,
      },
    } as IResource);
    const wrapper = shallow(
      <ApplicationStatusContainer
        store={store}
        deployRefs={[ref]}
        statefulsetRefs={[]}
        daemonsetRefs={[]}
      />,
    );
    const form = wrapper.find(ApplicationStatus);
    expect(form).toHaveProp({
      deployments: [item],
    });
  });
});
