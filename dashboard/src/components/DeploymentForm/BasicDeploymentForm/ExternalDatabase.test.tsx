import { shallow } from "enzyme";
import * as React from "react";

import ExternalDatabaseSection, {
  EXTERNAL_DB_HOST_PARAM_NAME,
  EXTERNAL_DB_PARAM_NAME,
  EXTERNAL_DB_PASSWORD_PARAM_NAME,
  EXTERNAL_DB_PORT_PARAM_NAME,
  EXTERNAL_DB_USER_PARAM_NAME,
  USE_SELF_HOSTED_DB_PARAM_NAME,
} from "./ExternalDatabase";

const defaultProps = {
  label: "Enable an external database",
  externalDatabaseParams: {
    [EXTERNAL_DB_PARAM_NAME]: { path: "edbs", value: {}, type: "object" },
    [EXTERNAL_DB_HOST_PARAM_NAME]: { path: "edbs.host", value: "localhost", type: "string" },
    [EXTERNAL_DB_USER_PARAM_NAME]: { path: "edbs.user", value: "user", type: "string" },
    [EXTERNAL_DB_PASSWORD_PARAM_NAME]: { path: "edbs.pass", value: "pass123", type: "string" },
    [EXTERNAL_DB_PORT_PARAM_NAME]: { path: "edbs.port", value: 1234, type: "integer" },
    [USE_SELF_HOSTED_DB_PARAM_NAME]: { path: "mariadb.enabled", value: true, type: "boolean" },
  },
  handleBasicFormParamChange: jest.fn(),
};

it("should render a external database section", () => {
  const wrapper = shallow(<ExternalDatabaseSection {...defaultProps} />);
  expect(wrapper).toMatchSnapshot();
});

it("should hide/show the database params if the self-hosted database is enabled/disabled", () => {
  const wrapper = shallow(<ExternalDatabaseSection {...defaultProps} />);
  expect(defaultProps.externalDatabaseParams[USE_SELF_HOSTED_DB_PARAM_NAME].value).toBe(true);
  expect(wrapper.find(".margin-t-normal").prop("hidden")).toBe(true);

  wrapper.setProps({
    ...defaultProps,
    externalDatabaseParams: {
      ...defaultProps.externalDatabaseParams,
      [USE_SELF_HOSTED_DB_PARAM_NAME]: { path: "mariadb.enabled", value: false, type: "boolean" },
    },
  });
  wrapper.update();
  expect(wrapper.find(".margin-t-normal").prop("hidden")).toBe(false);
});
