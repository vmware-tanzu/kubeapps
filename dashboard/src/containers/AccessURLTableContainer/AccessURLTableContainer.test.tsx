// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import { initialKinds } from "reducers/kube";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { keyForResourceRef } from "shared/ResourceRef";
import { ResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { IKubeItem, IKubeState, IResource, IStoreState } from "shared/types";
import AccessURLTableContainer from ".";
import AccessURLTable from "../../components/AppView/AccessURLTable";

const mockStore = configureMockStore([thunk]);

const makeStore = (resources: { [s: string]: IKubeItem<IResource> }) => {
  const state: IKubeState = {
    items: resources,
    kinds: initialKinds,
    subscriptions: {},
  };
  return mockStore({ kube: state, config: { featureFlags: {} } } as Partial<IStoreState>);
};

describe("AccessURLTableContainer", () => {
  it("maps Services and Ingresses in store to AccessURLTable props", () => {
    const ns = "wee";
    const name = "foo";
    const service = {
      isFetching: false,
      item: { metadata: { name: `${name}-service` } } as IResource,
    };
    const ingress = {
      isFetching: false,
      item: { metadata: { name: `${name}-ingress` } } as IResource,
    };
    const serviceRef = {
      apiVersion: "v1",
      kind: "Service",
      namespace: ns,
      name: `${name}-service`,
    } as ResourceRef;
    const serviceKey = keyForResourceRef(serviceRef);
    const ingressRef = {
      apiVersion: "v1",
      kind: "Ingress",
      namespace: ns,
      name: `${name}-ingress`,
    } as ResourceRef;
    const ingressKey = keyForResourceRef(ingressRef);
    const store = makeStore({
      [serviceKey]: service,
      [ingressKey]: ingress,
    });
    const wrapper = shallow(
      <AccessURLTableContainer
        store={store}
        serviceRefs={[serviceRef]}
        ingressRefs={[ingressRef]}
      />,
    );
    const component = wrapper.find(AccessURLTable);
    expect(component).toHaveProp({
      services: [service],
      ingresses: [ingress],
    });
  });
});
