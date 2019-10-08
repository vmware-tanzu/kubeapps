import { mount, shallow } from "enzyme";
import * as React from "react";

import { IBasicFormParam } from "shared/types";
import DatabaseSection, { IDatabaseSectionProps } from "./DatabaseSection"; // , {
import TextParam from "./TextParam";

const defaultProps = {
  label: "Enable an external database",
  param: {
    children: {
      externalDatabaseDB: {
        path: "externalDatabase.database",
        type: "string",
        value: "bitnami_wordpress",
      },
      externalDatabaseHost: { path: "externalDatabase.host", type: "string", value: "localhost" },
      externalDatabasePassword: { path: "externalDatabase.password", type: "string" },
      externalDatabasePort: { path: "externalDatabase.port", type: "integer", value: 3306 },
      externalDatabaseUser: {
        path: "externalDatabase.user",
        type: "string",
        value: "bn_wordpress",
      },
    },
    path: "externalDatabase",
    title: "External Database Details",
    type: "object",
  } as IBasicFormParam,
  disableExternalDBParam: {
    path: "mariadb.enabled",
    title: "Enable External Database",
    type: "boolean",
    value: true,
  } as IBasicFormParam,
  disableExternalDBParamName: "disableExternalDB",
  handleBasicFormParamChange: jest.fn(),
  appValues: "externalDatabase: {}",
  handleValuesChange: jest.fn(),
} as IDatabaseSectionProps;

it("should render a external database section", () => {
  const wrapper = shallow(<DatabaseSection {...defaultProps} />);
  expect(wrapper).toMatchSnapshot();
});

it("should hide/show the database params if the self-hosted database is enabled/disabled", () => {
  const wrapper = shallow(<DatabaseSection {...defaultProps} />);
  expect(defaultProps.disableExternalDBParam.value).toBe(true);
  expect(wrapper.find(".margin-t-normal").prop("hidden")).toBe(true);

  wrapper.setProps({
    ...defaultProps,
    disableExternalDBParam: { path: "mariadb.enabled", value: false, type: "boolean" },
  });
  wrapper.update();
  expect(wrapper.find(".margin-t-normal").prop("hidden")).toBe(false);
});

it("should change the parent parameter when a children is modified", () => {
  const wrapper = mount(<DatabaseSection {...defaultProps} />);

  const hostParam = wrapper.find(TextParam).findWhere(t => t.prop("label") === "Host");
  (hostParam.prop("handleBasicFormParamChange") as any)(
    "externalDatabaseHost",
    defaultProps.param.children!.externalDatabaseHost,
  )({ currentTarget: { value: "foo" } });

  expect(defaultProps.param.children!.externalDatabaseHost.value).toBe("foo");
});
