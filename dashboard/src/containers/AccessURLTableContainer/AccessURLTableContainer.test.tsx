import { shallow } from "enzyme";
import { initialKinds } from "reducers/kube";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import ResourceRef, { keyForResourceRef } from "shared/ResourceRef";
import { ResourceRef as APIResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { IKubeItem, IKubeState, IResource } from "shared/types";
import AccessURLTableContainer from ".";
import AccessURLTable from "../../components/AppView/AccessURLTable";

const mockStore = configureMockStore([thunk]);
const clusterName = "cluster-name";

const makeStore = (resources: { [s: string]: IKubeItem<IResource> }) => {
  const state: IKubeState = {
    items: resources,
    kinds: initialKinds,
    sockets: {},
    subscriptions: {},
    timers: {},
  };
  return mockStore({ kube: state, config: { featureFlags: {} } });
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
    const serviceRef = new ResourceRef(
      {
        apiVersion: "v1",
        kind: "Service",
        namespace: ns,
        name: `${name}-service`,
      } as APIResourceRef,
      clusterName,
      "services",
      true,
      "default",
    );
    const serviceKey = keyForResourceRef(
      serviceRef.apiVersion,
      serviceRef.kind,
      serviceRef.namespace,
      serviceRef.name,
    );
    const ingressRef = new ResourceRef(
      {
        apiVersion: "v1",
        kind: "Ingress",
        namespace: ns,
        name: `${name}-ingress`,
      } as APIResourceRef,
      clusterName,
      "ingresses",
      true,
      "default",
    );
    const ingressKey = keyForResourceRef(
      ingressRef.apiVersion,
      ingressRef.kind,
      ingressRef.namespace,
      ingressRef.name,
    );
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
