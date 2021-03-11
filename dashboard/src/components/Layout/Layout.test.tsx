import actions from "actions";
import * as ReactRedux from "react-redux";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import Layout from "./Layout";

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.namespace };
beforeEach(() => {
  actions.kube = {
    ...actions.kube,
    getResourceKinds: jest.fn(),
  };
  const mockDispatch = jest.fn(res => res);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.namespace = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
  jest.resetAllMocks();
});

const defaultState = {
  clusters: {
    currentCluster: "default",
    clusters: {
      default: {
        currentNamespace: "default",
        namespaces: ["default", "other"],
      },
    },
  },
  auth: { authenticated: true },
};

it("fetch resource kinds", () => {
  mountWrapper(getStore(defaultState), <Layout />);
  expect(actions.kube.getResourceKinds).toHaveBeenCalled();
});
