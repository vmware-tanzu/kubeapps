// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount } from "enzyme";
import { DeploymentEvent, IBasicFormParam } from "shared/types";
import BasicDeploymentForm, { IBasicDeploymentFormProps } from "./BasicDeploymentForm";

//TODO(agamez): enable back this test suite ASAP
/* eslint-disable jest/no-disabled-tests */

jest.useFakeTimers();

const defaultProps = {
  deploymentEvent: "install" as DeploymentEvent,
  handleBasicFormParamChange: jest.fn(() => jest.fn()),
  saveAllChanges: jest.fn(),
  isLoading: false,
  paramsFromComponentState: [],
} as IBasicDeploymentFormProps;

[
  {
    description: "renders a basic deployment with a username",
    params: [{ path: "wordpressUsername", value: "user" } as unknown as IBasicFormParam],
  },
  {
    description: "renders a basic deployment with a password",
    params: [{ path: "wordpressPassword", value: "sserpdrow" } as unknown as IBasicFormParam],
  },
  {
    description: "renders a basic deployment with a email",
    params: [{ path: "wordpressEmail", value: "user@example.com" } as unknown as IBasicFormParam],
  },
  {
    description: "renders a basic deployment with a generic string",
    params: [{ path: "blogName", value: "my-blog", type: "string" } as unknown as IBasicFormParam],
  },
  {
    description: "renders a basic deployment with custom configuration",
    params: [
      {
        path: "configuration",
        value: "First line\nSecond line",
        render: "textArea",
        type: "string",
      } as unknown as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a disk size",
    params: [
      {
        path: "size",
        value: "10Gi",
        type: "string",
        render: "slider",
      } as unknown as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a integer disk size",
    params: [
      {
        path: "size",
        value: 10,
        type: "integer",
        render: "slider",
      } as unknown as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a number disk size",
    params: [
      {
        path: "size",
        value: 10.0,
        type: "number",
        render: "slider",
      } as unknown as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with slider parameters",
    params: [
      {
        path: "size",
        value: "10Gi",
        type: "string",
        render: "slider",
        sliderMin: 1,
        sliderMax: 100,
        sliderStep: 1,
        sliderUnit: "Gi",
      } as unknown as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with username, password, email and a generic string",
    params: [
      { path: "wordpressUsername", value: "user" } as unknown as IBasicFormParam,
      { path: "wordpressPassword", value: "sserpdrow" } as unknown as IBasicFormParam,
      { path: "wordpressEmail", value: "user@example.com" } as unknown as IBasicFormParam,
      { path: "blogName", value: "my-blog", type: "string" } as unknown as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a generic boolean",
    params: [{ path: "enableMetrics", value: true, type: "boolean" } as unknown as IBasicFormParam],
  },
  {
    description: "renders a basic deployment with a generic number",
    params: [{ path: "replicas", value: 1, type: "integer" } as unknown as IBasicFormParam],
  },
].forEach(t => {
  it.skip(t.description, () => {
    const onChange = jest.fn();
    const handleBasicFormParamChange = jest.fn(() => onChange);
    const wrapper = mount(
      <BasicDeploymentForm
        {...defaultProps}
        paramsFromComponentState={t.params}
        handleBasicFormParamChange={handleBasicFormParamChange}
      />,
    );

    t.params.forEach((param, i) => {
      let input = wrapper.find(`input#${param.path}-${i}`);
      switch (param.type) {
        case "number":
        case "integer":
          if (param.render === "slider") {
            expect(wrapper.find("Slider")).toExist();
            break;
          }
          expect(input.prop("type")).toBe("number");
          break;
        case "string":
          if (param.render === "slider") {
            expect(wrapper.find("Slider")).toExist();
            break;
          }
          if (param.render === "textArea") {
            input = wrapper.find(`textarea#${param.path}-${i}`);
            expect(input).toExist();
            break;
          }
          if (param.path.match("Password")) {
            expect(input.prop("type")).toBe("password");
            break;
          }
          expect(input.prop("type")).toBe("string");
          break;
        default:
        // Ignore the rest of cases
      }
      input.simulate("change");
      const mockCalls = handleBasicFormParamChange.mock.calls;
      expect(mockCalls[i]).toEqual([param]);
      jest.runAllTimers();
      expect(onChange.mock.calls.length).toBe(i + 1);
    });
  });
});

it.skip("should render an external database section", () => {
  const params = [
    {
      path: "edbs",
      value: {},
      type: "object",
      children: [{ path: "mariadb.enabled", value: {}, type: "boolean" }],
    } as unknown as IBasicFormParam,
  ];
  const wrapper = mount(
    <BasicDeploymentForm {...defaultProps} paramsFromComponentState={params} />,
  );

  const dbsec = wrapper.find("Subsection");
  expect(dbsec).toExist();
});

it.skip("should hide an element if it depends on a param (string)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: "bar",
    },
    {
      path: "bar",
      type: "boolean",
    },
  ] as unknown as IBasicFormParam[];
  const wrapper = mount(
    <BasicDeploymentForm {...defaultProps} paramsFromComponentState={params} />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
});

it.skip("should hide an element if it depends on a single param (object)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: {
        value: "enabled",
        path: "bar",
      },
    },
    {
      path: "bar",
      type: "string",
    },
  ] as unknown as IBasicFormParam[];
  const wrapper = mount(
    <BasicDeploymentForm
      {...defaultProps}
      isLoading={false}
      paramsFromComponentState={params} //
    />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
});

it.skip("should hide an element using hidden path and values even if it is not present in values.yaml (simple)", () => {
  const params = [
    {
      default: "a",
      enum: ["a", "b"],
      path: "dropdown",
      type: "string",
      value: "a",
    },
    {
      hidden: { path: "dropdown", value: "b" },
      path: "a",
      type: "string",
    },
    {
      hidden: { path: "dropdown", value: "a" },
      path: "b",
      type: "string",
    },
  ] as unknown as IBasicFormParam[];
  const wrapper = mount(
    <BasicDeploymentForm {...defaultProps} paramsFromComponentState={params} />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
  expect(hiddenParam.text()).toBe("b");
});

it.skip("should hide an element using hidden path and values even if it is not present in values.yaml (different depth levels)", () => {
  const params = [
    {
      default: "a",
      enum: ["a", "b"],
      path: "dropdown",
      type: "string",
      value: "a",
    },
    {
      hidden: { path: "secondLevelProperties/2dropdown", value: "2b" },
      path: "a",
      type: "string",
    },
    {
      hidden: { path: "secondLevelProperties/2dropdown", value: "2a" },
      path: "b",
      type: "string",
    },
    {
      default: "2a",
      enum: ["2a", "2b"],
      path: "secondLevelProperties/2dropdown",
      type: "string",
      value: "2a",
    },
    {
      hidden: { path: "dropdown", value: "b" },
      path: "secondLevelProperties/2a",
      type: "string",
    },
    {
      hidden: { path: "dropdown", value: "a" },
      path: "secondLevelProperties/2b",
      type: "string",
    },
  ] as unknown as IBasicFormParam[];
  const wrapper = mount(
    <BasicDeploymentForm {...defaultProps} paramsFromComponentState={params} />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
  expect(hiddenParam.filterWhere(p => p.text().includes("b"))).toExist();
  expect(hiddenParam.filterWhere(p => p.text().includes("2b"))).toExist();
});

it.skip("should hide an element if it depends on multiple params (AND) (object)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: {
        conditions: [
          {
            value: "enabled",
            path: "bar",
          },
          {
            value: "disabled",
            path: "baz",
          },
        ],
        operator: "and",
      },
    },
    {
      path: "bar",
      type: "string",
    },
  ] as unknown as IBasicFormParam[];
  const wrapper = mount(
    <BasicDeploymentForm {...defaultProps} paramsFromComponentState={params} />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
});

it.skip("should hide an element if it depends on multiple params (OR) (object)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: {
        conditions: [
          {
            value: "enabled",
            path: "bar",
          },
          {
            value: "disabled",
            path: "baz",
          },
        ],
        operator: "or",
      },
    },
    {
      path: "bar",
      type: "string",
    },
  ] as unknown as IBasicFormParam[];
  const wrapper = mount(
    <BasicDeploymentForm {...defaultProps} paramsFromComponentState={params} />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
});

it.skip("should hide an element if it depends on multiple params (NOR) (object)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: {
        conditions: [
          {
            value: "enabled",
            path: "bar",
          },
          {
            value: "disabled",
            path: "baz",
          },
        ],
        operator: "nor",
      },
    },
    {
      path: "bar",
      type: "string",
    },
  ] as unknown as IBasicFormParam[];
  const wrapper = mount(
    <BasicDeploymentForm {...defaultProps} paramsFromComponentState={params} />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
});

it.skip("should hide an element if it depends on the deploymentEvent (install | upgrade) (object)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: {
        event: "upgrade",
      },
    },
  ] as unknown as IBasicFormParam[];
  const wrapper = mount(
    <BasicDeploymentForm
      {...defaultProps}
      deploymentEvent="upgrade"
      paramsFromComponentState={params}
    />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
});

it.skip("should NOT hide an element if it depends on the deploymentEvent (install | upgrade) (object)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: {
        event: "upgrade",
      },
    },
  ] as unknown as IBasicFormParam[];
  const wrapper = mount(
    <BasicDeploymentForm
      {...defaultProps}
      deploymentEvent="install"
      paramsFromComponentState={params}
    />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).not.toExist();
});

it.skip("should hide an element if it depends on deploymentEvent (install | upgrade) combined with multiple params (object)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: {
        conditions: [
          {
            event: "upgrade",
          },
          {
            value: "enabled",
            path: "bar",
          },
        ],
        operator: "or",
      },
    },
    {
      path: "bar",
      type: "string",
    },
  ] as unknown as IBasicFormParam[];
  const wrapper = mount(
    <BasicDeploymentForm
      {...defaultProps}
      deploymentEvent="upgrade"
      paramsFromComponentState={params}
    />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
});
