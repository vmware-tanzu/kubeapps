import { shallow } from "enzyme";
import * as React from "react";

import { UnexpectedErrorAlert } from ".";
import { IRBACRole } from "../../shared/types";
import ErrorAlertHeader from "./ErrorAlertHeader";
import PermissionsErrorAlert from "./PermissionsErrorAlert";
import PermissionsListItem from "./PermissionsListItem";
import { genericMessage } from "./UnexpectedErrorAlert";

it("renders an error message for the action", () => {
  const roles: IRBACRole[] = [];
  const action = "unit-test";
  const wrapper = shallow(
    <PermissionsErrorAlert roles={roles} action={action} namespace="test" rawMessage="" />,
  );
  const header = wrapper
    .find(UnexpectedErrorAlert)
    .shallow()
    .find(ErrorAlertHeader);
  expect(header).toExist();
  expect(header.shallow().text()).toContain(`You don't have sufficient permissions to ${action}`);
  expect(wrapper.html()).toContain("The following error was returned:");
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
  const wrapper = shallow(
    <PermissionsErrorAlert roles={roles} action="test" namespace="test" rawMessage="" />,
  );
  expect(wrapper.html()).toContain("Ask your administrator for the following RBAC roles:");
  expect(wrapper.find(PermissionsListItem)).toHaveLength(2);
});

it("renders a link to access control documentation", () => {
  const roles: IRBACRole[] = [];
  const wrapper = shallow(
    <PermissionsErrorAlert roles={roles} action="test" namespace="test" rawMessage="" />,
  );
  expect(wrapper.html()).toMatch(
    /See the documentation for more info on.*access control in Kubeapps./,
  );
  expect(wrapper.html()).toContain(
    '<a href="https://github.com/kubeapps/kubeapps/blob/master/docs/user/access-control.md" target="_blank" rel="noopener noreferrer">',
  );
  expect(wrapper.html()).not.toContain(shallow(genericMessage).html());
});

it("renders the raw message if there are no roles defined", () => {
  const roles: IRBACRole[] = [];
  const wrapper = shallow(
    <PermissionsErrorAlert roles={roles} action="test" namespace="test" rawMessage="boom!" />,
  );
  expect(wrapper.html()).toMatch(/boom!/);
  expect(wrapper.html()).not.toContain(shallow(genericMessage).html());
});
