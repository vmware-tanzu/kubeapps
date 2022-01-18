// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import { IBasicFormParam } from "shared/types";
import Subsection, { ISubsectionProps } from "./Subsection";

const defaultProps = {
  label: "Enable an external database",
  param: {
    children: [
      {
        path: "externalDatabase.database",
        type: "string",
        value: "bitnami_wordpress",
      },
      { path: "externalDatabase.host", type: "string", value: "localhost" },
      { path: "externalDatabase.password", type: "string" },
      { path: "externalDatabase.port", type: "integer", value: 3306 },
      {
        path: "externalDatabase.user",
        type: "string",
        value: "bn_wordpress",
      },
      {
        path: "mariadb.enabled",
        title: "Enable External Database",
        type: "boolean",
        value: true,
      } as IBasicFormParam,
    ],
    path: "externalDatabase",
    title: "External Database Details",
    description: "description of the param",
    type: "object",
  } as IBasicFormParam,
  allParams: [],
  appValues: "externalDatabase: {}",
  deploymentEvent: "install",
  handleValuesChange: jest.fn(),
} as ISubsectionProps;

it("should render a external database section", () => {
  const wrapper = shallow(<Subsection {...defaultProps} />);
  expect(wrapper).toMatchSnapshot();
});
