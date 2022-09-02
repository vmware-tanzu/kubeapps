// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import { ResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { initialKinds } from "reducers/kube";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { keyForResourceRef } from "shared/ResourceRef";
import { IKubeItem, IKubeState, IResource, IStoreState } from "shared/types";
import ApplicationStatusContainer from ".";
import ApplicationStatus from "../../components/ApplicationStatus";

const mockStore = configureMockStore([thunk]);

const makeStore = (resources: { [s: string]: IKubeItem<IResource> }) => {
  const state: IKubeState = {
    items: resources,
    kinds: initialKinds,
    subscriptions: {},
  };
  return mockStore({ kube: state, config: { featureFlags: {} } } as Partial<IStoreState>);
};

describe("ApplicationStatusContainer", () => {
  it("maps Deployment in store to ApplicationStatus props", () => {
    const ns = "wee";
    const name = "foo";
    const item = { isFetching: false, item: { metadata: { name } } as IResource };
    const ref = {
      apiVersion: "apps/v1",
      kind: "Deployment",
      namespace: ns,
      name,
    } as ResourceRef;
    const key = keyForResourceRef(ref);
    const store = makeStore({
      [key]: item,
    });
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
