import { shallow } from "enzyme";
import * as React from "react";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";

import { IKubeItem, IKubeState, IResource } from "shared/types";
import DeploymentItemContainer from ".";
import DeploymentItem from "../../components/AppView/DeploymentsTable/DeploymentItem";
import ResourceRef from "../../shared/ResourceRef";

const mockStore = configureMockStore([thunk]);

const makeStore = (secrets: { [s: string]: IKubeItem<IResource> }) => {
  const state: IKubeState = {
    items: secrets,
  };
  return mockStore({ kube: state });
};

describe("DeploymentItemContainer", () => {
  it("maps Deployment in store to DeploymentItem props", () => {
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
    const wrapper = shallow(<DeploymentItemContainer store={store} deployRef={ref} />);
    const form = wrapper.find(DeploymentItem);
    expect(form).toHaveProp({
      name,
      deployment: item,
    });
  });
});
