import actions from "actions";
import Table from "components/js/Table";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper.v2";
import * as React from "react";
import * as ReactRedux from "react-redux";
import ResourceRef from "shared/ResourceRef";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IResource } from "shared/types";
import ResourceTable from "./ResourceTable.v2";

const defaultProps = {
  id: "test",
  resourceRefs: [],
  watchResource: jest.fn(),
  closeWatch: jest.fn(),
};

const sampleResourceRef = {
  cluster: "cluster-name",
  apiVersion: "v1",
  kind: "Deployment",
  name: "foo",
  namespace: "default",
  filter: "",
  getResourceURL: jest.fn(() => "deployment-foo"),
  watchResourceURL: jest.fn(),
  getResource: jest.fn(),
  watchResource: jest.fn(),
} as ResourceRef;

const deployment = {
  metadata: {
    name: "foo",
  },
  status: {
    replicas: 1,
    updatedReplicas: 0,
    availableReplicas: 0,
  },
};

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.kube = {
    ...actions.kube,
    getAndWatchResource: jest.fn(),
    closeWatchResource: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("watches the given resources and close watchers", async () => {
  const watchResource = jest.fn();
  const closeWatch = jest.fn();
  actions.kube = {
    ...actions.kube,
    getAndWatchResource: watchResource,
    closeWatchResource: closeWatch,
  };
  const wrapper = mountWrapper(
    defaultStore,
    <ResourceTable {...defaultProps} resourceRefs={[sampleResourceRef]} />,
  );
  expect(watchResource).toHaveBeenCalledWith(sampleResourceRef);
  wrapper.unmount();
  expect(closeWatch).toHaveBeenCalledWith(sampleResourceRef);
});

it("renders a table with a resource", () => {
  const state = getStore({
    kube: {
      items: {
        "deployment-foo": {
          isFetching: false,
          item: deployment as IResource,
        },
      },
    },
  });
  const wrapper = mountWrapper(
    state,
    <ResourceTable {...defaultProps} resourceRefs={[sampleResourceRef]} />,
  );
  expect(wrapper.find(Table).prop("data")).toEqual([
    { name: "foo", desired: 1, upToDate: 0, available: 0 },
  ]);
});

it("renders a table with a loading resource", () => {
  const state = getStore({
    kube: {
      items: {
        "deployment-foo": {
          isFetching: true,
        },
      },
    },
  });
  const wrapper = mountWrapper(
    state,
    <ResourceTable {...defaultProps} resourceRefs={[sampleResourceRef]} />,
  );

  const data = wrapper.find(Table).prop("data");
  const row = data[0];
  expect(row.name).toEqual("foo");
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("renders a table with an error", () => {
  const state = getStore({
    kube: {
      items: {
        "deployment-foo": {
          isFetching: false,
          error: new Error("Boom!"),
        },
      },
    },
  });
  const wrapper = mountWrapper(
    state,
    <ResourceTable {...defaultProps} resourceRefs={[sampleResourceRef]} />,
  );

  const data = wrapper.find(Table).prop("data");
  const row = data[0];
  expect(row.name).toEqual("foo");
  expect(wrapper.text()).toContain("Error: Boom!");
});
