import { shallow } from "enzyme";
import * as React from "react";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";

import { IKubeItem, IKubeState, IResource } from "shared/types";
import AccessURLTableContainer from ".";
import AccessURLTable from "../../components/AppView/AccessURLTable";
import ResourceRef from "../../shared/ResourceRef";

const mockStore = configureMockStore([thunk]);

const makeStore = (resources: { [s: string]: IKubeItem<IResource> }) => {
  const state: IKubeState = {
    items: resources,
    sockets: {},
  };
  return mockStore({ kube: state });
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
    const store = makeStore({
      "api/kube/api/v1/namespaces/wee/services/foo-service": service,
      "api/kube/api/v1/namespaces/wee/ingresses/foo-ingress": ingress,
    });
    const serviceRef = new ResourceRef({
      apiVersion: "v1",
      kind: "Service",
      metadata: {
        namespace: ns,
        name: `${name}-service`,
      },
    } as IResource);
    const ingressRef = new ResourceRef({
      apiVersion: "v1",
      kind: "Ingress",
      metadata: {
        namespace: ns,
        name: `${name}-ingress`,
      },
    } as IResource);
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
