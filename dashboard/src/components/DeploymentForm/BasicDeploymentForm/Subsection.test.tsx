import { mount, shallow } from "enzyme";
import * as React from "react";

import { IBasicFormParam } from "shared/types";
import Subsection, { ISubsectionProps } from "./Subsection";
import TextParam from "./TextParam";

const defaultProps = {
  label: "Enable an external database",
  name: "externalDatabase",
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
      useSelfHostedDatabase: {
        path: "mariadb.enabled",
        title: "Enable External Database",
        type: "boolean",
        value: true,
      } as IBasicFormParam,
    },
    path: "externalDatabase",
    title: "External Database Details",
    type: "object",
  } as IBasicFormParam,
  handleBasicFormParamChange: jest.fn(),
  appValues: "externalDatabase: {}",
  handleValuesChange: jest.fn(),
  enablerChildrenParam: "useSelfHostedDatabase",
  enablerCondition: false,
} as ISubsectionProps;

it("should render a external database section", () => {
  const wrapper = shallow(<Subsection {...defaultProps} />);
  expect(wrapper).toMatchSnapshot();
});

it("should hide/show the database params if the self-hosted database is enabled/disabled", () => {
  const wrapper = shallow(<Subsection {...defaultProps} />);
  expect(defaultProps.param.children!.useSelfHostedDatabase.value).toBe(true);
  expect(wrapper.find(".margin-t-normal").prop("hidden")).toBe(true);

  wrapper.setProps({
    ...defaultProps,
    param: {
      ...defaultProps.param,
      children: {
        ...defaultProps.param.children,
        useSelfHostedDatabase: { path: "mariadb.enabled", value: false, type: "boolean" },
      },
    },
  });
  wrapper.update();
  expect(wrapper.find(".margin-t-normal").prop("hidden")).toBe(false);
});

it("should change the parent parameter when a children is modified", () => {
  const wrapper = mount(<Subsection {...defaultProps} />);

  const hostParam = wrapper.find(TextParam).findWhere(t => t.prop("label") === "Host");
  (hostParam.prop("handleBasicFormParamChange") as any)(
    "externalDatabaseHost",
    defaultProps.param.children!.externalDatabaseHost,
  )({ currentTarget: { value: "foo" } });

  expect(defaultProps.param.children!.externalDatabaseHost.value).toBe("foo");
});
