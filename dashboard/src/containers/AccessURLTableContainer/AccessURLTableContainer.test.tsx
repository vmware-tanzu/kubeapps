import { shallow } from "enzyme";
import { initialKinds } from "reducers/kube";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";

import { IKubeItem, IKubeState, IResource } from "shared/types";
import AccessURLTableContainer from ".";
import AccessURLTable from "../../components/AppView/AccessURLTable";
import ResourceRef from "../../shared/ResourceRef";

const mockStore = configureMockStore([thunk]);
const clusterName = "cluster-name";

const makeStore = (resources: { [s: string]: IKubeItem<IResource> }) => {
  const state: IKubeState = {
    items: resources,
    kinds: initialKinds,
    sockets: {},
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
    const store = makeStore({
      [`api/clusters/${clusterName}/api/v1/namespaces/wee/services/foo-service`]: service,
      [`api/clusters/${clusterName}/api/v1/namespaces/wee/ingresses/foo-ingress`]: ingress,
    });
    const serviceRef = new ResourceRef(
      {
        apiVersion: "v1",
        kind: "Service",
        metadata: {
          namespace: ns,
          name: `${name}-service`,
        },
      } as IResource,
      clusterName,
      "services",
      true,
    );
    const ingressRef = new ResourceRef(
      {
        apiVersion: "v1",
        kind: "Ingress",
        metadata: {
          namespace: ns,
          name: `${name}-ingress`,
        },
      } as IResource,
      clusterName,
      "ingresses",
      true,
    );
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
