import Table from "components/js/Table";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";

import ResourceRef from "shared/ResourceRef";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IResource } from "shared/types";
import ResourceTable from "./ResourceTable";

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
  plural: "deployments",
  namespaced: true,
  getResourceURL: jest.fn(() => "deployment-foo"),
  watchResourceURL: jest.fn(),
  getResource: jest.fn(),
  watchResource: jest.fn(),
} as ResourceRef;

const deployment = {
  kind: "Deployment",
  metadata: {
    name: "foo",
  },
  status: {
    replicas: 1,
    updatedReplicas: 0,
    availableReplicas: 0,
  },
};

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

it("do not fail if the resources are already populated but the refs not yet", () => {
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
  const wrapper = mountWrapper(state, <ResourceTable {...defaultProps} resourceRefs={[]} />);
  expect(wrapper.find(Table)).not.toExist();
});

it("renders the table with two entries with a list reference", () => {
  const state = getStore({
    kube: {
      items: {
        "deployment-foo": {
          isFetching: false,
          item: {
            items: [
              deployment as IResource,
              { ...deployment, metadata: { name: "bar" } } as IResource,
            ],
          },
        },
      },
    },
  });
  const wrapper = mountWrapper(
    state,
    <ResourceTable
      {...defaultProps}
      resourceRefs={[{ ...sampleResourceRef, name: "" } as ResourceRef]}
    />,
  );
  expect(wrapper.find(Table).prop("data")).toHaveLength(2);
});
