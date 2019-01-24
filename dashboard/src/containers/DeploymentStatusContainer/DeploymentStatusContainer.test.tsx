import { shallow } from "enzyme";
import * as React from "react";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";

import DeploymentStatusContainer from ".";
import DeploymentStatus from "../../components/DeploymentStatus";
import ResourceRef from "../../shared/ResourceRef";
import { IKubeItem, IKubeState, IResource } from "../../shared/types";

const mockStore = configureMockStore([thunk]);

const makeStore = (resources: { [s: string]: IKubeItem<IResource> }) => {
  const state: IKubeState = {
    items: resources,
  };
  return mockStore({ kube: state });
};

describe("DeploymentStatusContainer", () => {
  it("maps Deployment in store to DeploymentStatus props", () => {
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
    const wrapper = shallow(<DeploymentStatusContainer store={store} deployRefs={[ref]} />);
    const form = wrapper.find(DeploymentStatus);
    expect(form).toHaveProp({
      deployments: [item],
    });
  });
});
