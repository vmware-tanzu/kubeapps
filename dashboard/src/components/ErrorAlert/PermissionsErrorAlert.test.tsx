import { shallow } from "enzyme";
import * as React from "react";

import { IRBACRole } from "../../shared/types";
import ErrorAlertHeader from "./ErrorAlertHeader";
import PermissionsErrorAlert from "./PermissionsErrorAlert";
import PermissionsListItem from "./PermissionsListItem";

it("renders an error message for the action", () => {
  const roles: IRBACRole[] = [];
  const action = "unit-test";
  const wrapper = shallow(<PermissionsErrorAlert roles={roles} action={action} namespace="test" />);
  expect(
    (wrapper.find(ErrorAlertHeader).props().children as Array<string | JSX.Element>).join(""),
  ).toContain(`You don't have sufficient permissions to ${action}`);
  expect(wrapper.text()).toContain("Ask your administrator for the following RBAC roles:");
  expect(wrapper).toMatchSnapshot();
});

it("renders PermissionsListItem for each RBAC role", () => {
  const roles: IRBACRole[] = [
    {
      apiGroup: "test.kubeapps.com",
      resource: "tests",
      verbs: ["get", "create"],
    },
    {
      apiGroup: "apps",
      namespace: "test",
      resource: "deployments",
      verbs: ["list", "watch"],
    },
  ];
  const wrapper = shallow(<PermissionsErrorAlert roles={roles} action="test" namespace="test" />);
  expect(wrapper.find(PermissionsListItem)).toHaveLength(2);
});

it("renders a link to access control documentation", () => {
  const roles: IRBACRole[] = [];
  const wrapper = shallow(<PermissionsErrorAlert roles={roles} action="test" namespace="test" />);
  expect(wrapper.text()).toContain(
    "See the documentation for more info on access control in Kubeapps.",
  );
  expect(wrapper.find("a").props()).toMatchObject({
    href: "https://github.com/kubeapps/kubeapps/blob/master/docs/user/access-control.md",
    target: "_blank",
  });
});
