import * as React from "react";

import { mount } from "enzyme";
import { IBasicFormParam } from "shared/types";
import BasicDeploymentForm from "./BasicDeploymentForm";
import Subsection from "./Subsection";

const defaultProps = {
  params: [],
  handleBasicFormParamChange: jest.fn(() => jest.fn()),
  appValues: "",
  handleValuesChange: jest.fn(),
};

[
  {
    description: "renders a basic deployment with a username",
    params: { username: { path: "wordpressUsername", value: "user" } as IBasicFormParam },
  },
  {
    description: "renders a basic deployment with a password",
    params: {
      password: { path: "wordpressPassword", value: "sserpdrow" } as IBasicFormParam,
    },
  },
  {
    description: "renders a basic deployment with a email",
    params: { email: { path: "wordpressEmail", value: "user@example.com" } as IBasicFormParam },
  },
  {
    description: "renders a basic deployment with a generic string",
    params: {
      blogName: { path: "blogName", value: "my-blog", type: "string" } as IBasicFormParam,
    },
  },
  {
    description: "renders a basic deployment with a disk size",
    params: {
      diskSize: { path: "size", value: "10Gi", type: "string" } as IBasicFormParam,
    },
  },
  {
    description: "renders a basic deployment with username, password, email and a generic string",
    params: {
      username: { path: "wordpressUsername", value: "user" } as IBasicFormParam,
      password: { path: "wordpressPassword", value: "sserpdrow" } as IBasicFormParam,
      email: { path: "wordpressEmail", value: "user@example.com" } as IBasicFormParam,
      blogName: { path: "blogName", value: "my-blog", type: "string" } as IBasicFormParam,
    },
  },
  {
    description: "renders a basic deployment with a generic boolean",
    params: {
      enableMetrics: { path: "enableMetrics", value: true, type: "boolean" } as IBasicFormParam,
    },
  },
  {
    description: "renders a basic deployment with a generic number",
    params: {
      replicas: { path: "replicas", value: 1, type: "integer" } as IBasicFormParam,
    },
  },
].forEach(t => {
  it(t.description, () => {
    const onChange = jest.fn();
    const handleBasicFormParamChange = jest.fn(() => onChange);
    const wrapper = mount(
      <BasicDeploymentForm
        {...defaultProps}
        params={t.params as any}
        handleBasicFormParamChange={handleBasicFormParamChange}
      />,
    );
    expect(wrapper).toMatchSnapshot();
    Object.keys(t.params).map((param, i) => {
      wrapper.find(`input#${param}-${i}`).simulate("change");
      const mockCalls = handleBasicFormParamChange.mock.calls;
      expect(mockCalls[i]).toEqual([param, t.params[param]]);
      expect(onChange.mock.calls.length).toBe(i + 1);
    });
  });
});

it("should render an external database section", () => {
  const params = {
    externalDatabase: {
      path: "edbs",
      value: {},
      type: "object",
      children: {
        useSelfHostedDatabase: { path: "mariadb.enabled", value: {}, type: "boolean" },
      },
    },
  };
  const wrapper = mount(<BasicDeploymentForm {...defaultProps} params={params} />);

  const dbsec = wrapper.find(Subsection);
  expect(dbsec).toExist();
});
