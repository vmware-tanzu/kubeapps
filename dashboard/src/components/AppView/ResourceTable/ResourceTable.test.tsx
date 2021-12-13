import Table from "components/js/Table";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import { keyForResourceRef } from "shared/ResourceRef";
import { ResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IResource } from "shared/types";
import ResourceTable from "./ResourceTable";

const defaultProps = {
  id: "test",
  resourceRefs: [],
};

const sampleResourceRef = {
  apiVersion: "v1",
  kind: "Deployment",
  name: "foo",
  namespace: "default",
} as ResourceRef;

const sampleKey = keyForResourceRef(sampleResourceRef);

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
        [sampleKey]: {
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
        [sampleKey]: {
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
        [sampleKey]: {
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
        [sampleKey]: {
          isFetching: false,
          item: deployment as IResource,
        },
      },
    },
  });
  const wrapper = mountWrapper(state, <ResourceTable {...defaultProps} resourceRefs={[]} />);
  expect(wrapper.find(Table)).not.toExist();
});
