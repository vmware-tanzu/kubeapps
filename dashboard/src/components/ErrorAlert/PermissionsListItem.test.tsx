import { shallow } from "enzyme";
import * as React from "react";

import { IRBACRole } from "../../shared/types";
import PermissionsListItem from "./PermissionsListItem";

it("renders list item for the role", () => {
  const role: IRBACRole = {
    apiGroup: "test.kubeapps.com",
    resource: "tests",
    verbs: ["get", "create"],
  };
  const wrapper = shallow(<PermissionsListItem role={role} namespace="test" />);
  expect(wrapper.text()).toContain("get, create tests (test.kubeapps.com) in the test namespace.");
  expect(wrapper).toMatchSnapshot();
});

it("renders the special case for cluster-wide roles", () => {
  const role: IRBACRole = {
    apiGroup: "test.kubeapps.com",
    clusterWide: true,
    resource: "tests",
    verbs: ["get", "create"],
  };
  const wrapper = shallow(<PermissionsListItem role={role} namespace="test" />);
  expect(wrapper.text()).toContain("get, create tests (test.kubeapps.com) in all namespaces.");
  expect(wrapper).toMatchSnapshot();
});

it("uses the roles' namespace over the current namespace if defined", () => {
  const role: IRBACRole = {
    apiGroup: "test.kubeapps.com",
    namespace: "another",
    resource: "tests",
    verbs: ["get", "create"],
  };
  const wrapper = shallow(<PermissionsListItem role={role} namespace="test" />);
  expect(wrapper.text()).toContain(
    "get, create tests (test.kubeapps.com) in the another namespace.",
  );
  expect(wrapper).toMatchSnapshot();
});
