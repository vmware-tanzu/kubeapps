import { shallow } from "enzyme";
import { initialKinds } from "reducers/kube";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import ResourceRef, { keyForResourceRef } from "shared/ResourceRef";
import { IKubeItem, IKubeState, IResource } from "shared/types";
import ApplicationStatusContainer from ".";
import ApplicationStatus from "../../components/ApplicationStatus";
import { ResourceRef as APIResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages";

const mockStore = configureMockStore([thunk]);
const clusterName = "cluster-Name";

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

describe("ApplicationStatusContainer", () => {
  it("maps Deployment in store to ApplicationStatus props", () => {
    const ns = "wee";
    const name = "foo";
    const item = { isFetching: false, item: { metadata: { name } } as IResource };
    const ref = new ResourceRef(
      {
        apiVersion: "apps/v1",
        kind: "Deployment",
        namespace: ns,
        name,
      } as APIResourceRef,
      clusterName,
      "deployments",
      true,
      "default",
    );
    const key = keyForResourceRef(ref.apiVersion, ref.kind, ref.namespace, ref.name);
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
