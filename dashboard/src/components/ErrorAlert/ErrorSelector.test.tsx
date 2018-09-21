import { shallow } from "enzyme";
import * as React from "react";

import { PermissionsErrorAlert } from ".";
import { definedNamespaces } from "../../shared/Namespace";
import {
  AppConflict,
  ForbiddenError,
  IRBACRole,
  NotFoundError,
  UnprocessableEntity,
} from "../../shared/types";
import ErrorPageHeader from "./ErrorAlertHeader";
import ErrorSelector from "./ErrorSelector";
import PermissionsListItem from "./PermissionsListItem";
import UnexpectedErrorAlert from "./UnexpectedErrorAlert";

describe("AppConflict", () => {
  it("should render a simple message with the name of the resource", () => {
    const wrapper = shallow(<ErrorSelector error={new AppConflict()} resource={"my app"} />);
    const errAlert = wrapper.find(UnexpectedErrorAlert);
    expect(errAlert).toExist();
    expect(errAlert.html()).toContain("my app already exists, try a different name");
    expect(errAlert.html()).not.toContain("Troubleshooting");
    expect(wrapper).toMatchSnapshot();
  });
});

describe("ForbiddenError", () => {
  it("should show an error message with the default RBAC roles", () => {
    const defaultRBACRoles = {
      view: [
        {
          apiGroup: "v1",
          namespace: "my-ns",
          resource: "my-app",
          verbs: ["get", "list"],
        } as IRBACRole,
      ],
    };
    const wrapper = shallow(
      <ErrorSelector
        error={new ForbiddenError()}
        resource={"my app"}
        namespace={"my-ns"}
        defaultRequiredRBACRoles={defaultRBACRoles}
        action="view"
      />,
    );
    const errAlert = wrapper.find(PermissionsErrorAlert);
    expect(errAlert).toExist();
    expect(errAlert.html()).not.toContain("Troubleshooting");
    const header = errAlert
      .shallow()
      .find(UnexpectedErrorAlert)
      .shallow()
      .find(ErrorPageHeader)
      .shallow();
    expect(header.text()).toContain(
      "You don't have sufficient permissions to view my app in the my-ns namespace",
    );
    expect(wrapper).toMatchSnapshot();
  });
  it("should extract the required RBAC roles from the error message", () => {
    const role = { apiGroup: "v1", namespace: "my-ns", resource: "my-app", verbs: ["get", "list"] };
    const message = JSON.stringify([role]);
    const wrapper = shallow(
      <ErrorSelector error={new ForbiddenError(message)} resource="my-app" />,
    );
    const items = wrapper
      .find(PermissionsErrorAlert)
      .shallow()
      .find(UnexpectedErrorAlert)
      .shallow()
      .find(PermissionsListItem);
    expect(items.length).toBe(1);
    expect(items.prop("role")).toMatchObject(role);
  });
});

describe("NotFoundError", () => {
  it("should show a not found error message", () => {
    const wrapper = shallow(<ErrorSelector error={new NotFoundError()} resource="my-app" />);
    expect(wrapper.html()).toContain("my-app not found");
    expect(wrapper.html()).not.toContain("Troubleshooting");
  });
  it("should include the namespace in the error if given", () => {
    const wrapper = shallow(
      <ErrorSelector error={new NotFoundError()} resource="my-app" namespace="my-ns" />,
    );
    expect(wrapper.html()).toMatch(/my-app not found.*in.*my-ns.*namespace/);
    expect(wrapper).toMatchSnapshot();
  });
  it("should include a warning if all-namespaces is selected", () => {
    const wrapper = shallow(
      <ErrorSelector
        error={new NotFoundError()}
        resource="my-app"
        namespace={definedNamespaces.all}
      />,
    );
    expect(wrapper.html()).toContain("You may need to select a namespace");
  });
});

describe("UnprocessableEntity", () => {
  it("Should show the text of the error message", () => {
    const wrapper = shallow(
      <ErrorSelector error={new UnprocessableEntity("that is wrong!")} resource="my-app" />,
    );
    expect(wrapper.html()).toContain("Sorry! Something went wrong processing my-app");
    expect(wrapper.html()).toContain("that is wrong!");
    expect(wrapper).toMatchSnapshot();
  });
});
