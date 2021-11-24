import { shallow } from "enzyme";
import { initialKinds } from "reducers/kube";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import ResourceRef from "shared/ResourceRef";
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
    timers: {},
  };
  return mockStore({ kube: state, config: { featureFlags: {} } });
};

describe("ApplicationStatusContainer", () => {
  it("maps Deployment in store to ApplicationStatus props", () => {
    const ns = "wee";
    const name = "foo";
    const item = { isFetching: false, item: { metadata: { name } } as IResource };
    const store = makeStore({
      [`api/clusters/${clusterName}/apis/apps/v1/namespaces/wee/deployments/foo`]: item,
    });
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
