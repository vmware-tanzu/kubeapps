import { mount, shallow } from "enzyme";
import * as React from "react";

import { IBasicFormParam } from "shared/types";
import BooleanParam from "./BooleanParam";
import Subsection, { ISubsectionProps } from "./Subsection";

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
        disables: "externalDatabase",
      } as IBasicFormParam,
    },
    path: "externalDatabase",
    title: "External Database Details",
    description: "description of the param",
    type: "object",
  } as IBasicFormParam,
  handleBasicFormParamChange: jest.fn(),
  appValues: "externalDatabase: {}",
  handleValuesChange: jest.fn(),
  enablerChildrenParam: "useSelfHostedDatabase",
  enablerCondition: false,
  renderParam: jest.fn(),
} as ISubsectionProps;

it("should render a external database section", () => {
  const wrapper = shallow(<Subsection {...defaultProps} />);
  expect(wrapper).toMatchSnapshot();
});

it("should hide/show the database params if the self-hosted database is enabled/disabled", () => {
  const wrapper = shallow(<Subsection {...defaultProps} />);
  expect(defaultProps.param.children!.useSelfHostedDatabase.value).toBe(true);
  expect(wrapper.find("div").findWhere(d => d.prop("hidden"))).toExist();

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
  expect(wrapper.find("div").findWhere(d => d.prop("hidden"))).not.toExist();
});

it("should omit the enabler param if it doesn't exist", () => {
  const props = {
    ...defaultProps,
    param: {
      ...defaultProps.param,
      children: {
        ...defaultProps.param.children,
        useSelfHostedDatabase: {} as IBasicFormParam,
      },
    },
  };
  const wrapper = mount(<Subsection {...props} />);
  expect(wrapper.find(BooleanParam)).not.toExist();
  expect(wrapper.find("div").findWhere(d => d.prop("hidden"))).not.toExist();
});
