import { mount, shallow } from "enzyme";
import * as React from "react";

import { Tab, Tabs } from "react-tabs";
import itBehavesLike from "../../shared/specs";
import { IChartState, IChartVersion, NotFoundError } from "../../shared/types";
import ConfirmDialog from "../ConfirmDialog";
import { ErrorSelector } from "../ErrorAlert";
import ErrorPageHeader from "../ErrorAlert/ErrorAlertHeader";
import LoadingWrapper from "../LoadingWrapper";
import BasicDeploymentForm from "./BasicDeploymentForm/BasicDeploymentForm";
import DeploymentFormBody, {
  IDeploymentFormBodyProps,
  IDeploymentFormBodyState,
} from "./DeploymentFormBody";

const defaultProps = {
  chartID: "foo",
  chartVersion: "1.0.0",
  error: undefined,
  releaseName: undefined,
  selected: {} as IChartState["selected"],
  deployChart: jest.fn(),
  push: jest.fn(),
  fetchChartVersions: jest.fn(),
  getChartVersion: jest.fn(),
  namespace: "default",
  appValues: "foo: bar",
  valuesModified: false,
  setValues: jest.fn(),
  setValuesModified: jest.fn(),
  deployedValues: undefined,
} as IDeploymentFormBodyProps;
const versions = [{ id: "foo", attributes: { version: "1.2.3" } }] as IChartVersion[];

itBehavesLike("aLoadingComponent", { component: DeploymentFormBody, props: defaultProps });

describe("renders an error", () => {
  it("renders an error if it cannot find the given chart", () => {
    const wrapper = mount(
      <DeploymentFormBody
        {...defaultProps}
        selected={{ error: new NotFoundError() } as IChartState["selected"]}
      />,
    );
    expect(wrapper.find(ErrorPageHeader).exists()).toBe(true);
    expect(wrapper.find(ErrorPageHeader).text()).toContain('Chart "foo" (1.0.0) not found');
  });

  it("renders a generic error", () => {
    const wrapper = shallow(
      <DeploymentFormBody
        {...defaultProps}
        selected={{ error: new Error() } as IChartState["selected"]}
      />,
    );
    expect(wrapper.find(ErrorSelector).exists()).toBe(true);
    expect(wrapper.find(ErrorSelector).html()).toContain("Sorry! Something went wrong");
  });
});

it("renders the full DeploymentFormBody", () => {
  const wrapper = shallow(
    <DeploymentFormBody {...defaultProps} selected={{ versions, version: versions[0] }} />,
  );
  expect(wrapper).toMatchSnapshot();
});

const initialValues = "foo: bar";
const initialSchema = { properties: { foo: { type: "string", form: true } } };
const chartVersion = {
  id: "foo",
  attributes: { version: "1.0.0", app_version: "1.0", created: "1" },
  relationships: {
    chart: {
      data: {
        name: "chart",
        description: "chart-description",
        keywords: [],
        maintainers: [],
        repo: {
          name: "repo",
          url: "http://example.com",
        },
        sources: [],
      },
    },
  },
};
const props: IDeploymentFormBodyProps = {
  ...defaultProps,
  selected: {
    ...defaultProps.selected,
    versions: [chartVersion],
    version: chartVersion,
    values: initialValues,
    schema: initialSchema,
  },
};

describe("when there are changes in the selected version", () => {
  it("initializes the local values from props when props set", () => {
    const wrapper = shallow(<DeploymentFormBody {...defaultProps} />);
    wrapper.setProps({ selected: props.selected });
    const localState: IDeploymentFormBodyState = wrapper.instance()
      .state as IDeploymentFormBodyState;
    const basicFormParameters = [
      {
        form: true,
        path: "foo",
        value: "bar",
        type: "string",
      },
    ];
    expect(localState.basicFormParameters).toEqual(basicFormParameters);
  });

  describe("when the user has not modified any value", () => {
    it("selects the original values if the version doesn't change", () => {
      const setValues = jest.fn();
      const deployedValues = "foo: notBar";
      const wrapper = shallow(
        <DeploymentFormBody {...props} setValues={setValues} deployedValues={deployedValues} />,
      );
      wrapper.setProps({
        selected: {
          ...props.selected,
          values: "foo: ignored-value",
        },
      });
      const basicFormParameters = [
        {
          form: true,
          path: "foo",
          value: "notBar",
          type: "string",
        },
      ];
      const localState: IDeploymentFormBodyState = wrapper.instance()
        .state as IDeploymentFormBodyState;
      expect(localState.basicFormParameters).toEqual(basicFormParameters);
      expect(setValues).toHaveBeenCalledWith("foo: notBar\n");
    });

    it("uses the chart default values when original values are not defined", () => {
      const setValues = jest.fn();
      const wrapper = shallow(
        <DeploymentFormBody {...props} setValues={setValues} deployedValues={undefined} />,
      );
      wrapper.setProps({
        selected: {
          ...props.selected,
          values: "foo: notBar",
        },
      });
      const basicFormParameters = [
        {
          form: true,
          path: "foo",
          value: "notBar",
          type: "string",
        },
      ];
      const localState: IDeploymentFormBodyState = wrapper.instance()
        .state as IDeploymentFormBodyState;
      expect(localState.basicFormParameters).toEqual(basicFormParameters);
      expect(setValues).toHaveBeenCalledWith("foo: notBar");
    });
  });

  describe("when the user has modified the values", () => {
    it("will ignore original or default values", () => {
      const setValues = jest.fn();
      const deployedValues = "foo: ignored-value";
      const modifiedValues = "foo: notBar";
      const wrapper = shallow(
        <DeploymentFormBody
          {...props}
          setValues={setValues}
          deployedValues={deployedValues}
          valuesModified={true}
          appValues={modifiedValues}
        />,
      );
      wrapper.setProps({
        selected: {
          ...props.selected,
          values: "foo: another-ignored-value",
        },
      });
      const basicFormParameters = [
        {
          form: true,
          path: "foo",
          value: "notBar",
          type: "string",
        },
      ];
      const localState: IDeploymentFormBodyState = wrapper.instance()
        .state as IDeploymentFormBodyState;
      expect(localState.basicFormParameters).toEqual(basicFormParameters);
      expect(setValues).not.toHaveBeenCalled();
    });
  });
});

describe("when the basic form is enabled", () => {
  it("renders the different tabs", () => {
    const basicFormParameters = [
      {
        path: "wordpressUsername",
        value: "user",
      },
    ];
    const wrapper = mount(<DeploymentFormBody {...props} />);
    wrapper.setState({ appValues: "wordpressUsername: user", basicFormParameters });
    wrapper.update();
    expect(wrapper.find(LoadingWrapper)).not.toExist();
    expect(wrapper.find(Tabs)).toExist();
  });

  it("should not render the tabs if there are no basic parameters", () => {
    const wrapper = shallow(<DeploymentFormBody {...props} />);
    expect(wrapper.find(LoadingWrapper)).not.toExist();
    expect(wrapper.find(Tabs)).not.toExist();
  });

  it("changes the parameter value", () => {
    const basicFormParameters = [
      {
        path: "wordpressUsername",
        value: "user",
      },
    ];
    const setValuesModified = jest.fn();
    const setValues = jest.fn();
    const wrapper = mount(
      <DeploymentFormBody
        {...props}
        appValues="wordpressUsername: user"
        setValues={setValues}
        setValuesModified={setValuesModified}
      />,
    );
    wrapper.setState({ basicFormParameters });
    wrapper.update();

    // Fake onChange
    const input = wrapper.find(BasicDeploymentForm).find("input");
    const onChange = input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void;
    onChange({ currentTarget: { value: "foo" } } as React.FormEvent<HTMLInputElement>);

    expect(wrapper.state("basicFormParameters")).toEqual([
      { path: "wordpressUsername", value: "foo" },
    ]);
    expect(setValuesModified).toHaveBeenCalled();
    expect(setValues).toHaveBeenCalledWith("wordpressUsername: foo\n");
  });

  it("should update existing params if the app values change and the user clicks on the Basic tab", () => {
    const testProps = {
      ...props,
      selected: {
        ...props.selected,
        schema: { properties: { wordpressUsername: { type: "string", form: true } } },
      },
    };
    const basicFormParameters = [
      {
        path: "wordpressUsername",
        value: "user",
      },
    ];
    const wrapper = mount(<DeploymentFormBody {...testProps} />);
    wrapper.setState({ basicFormParameters });
    wrapper.setProps({ appValues: "wordpressUsername: foo" });
    wrapper.update();

    const tab = wrapper
      .find(Tab)
      .findWhere(t => !!t.text().match("Basic"))
      .first();
    tab.simulate("click");

    expect(wrapper.state("basicFormParameters")).toMatchObject([
      {
        path: "wordpressUsername",
        value: "foo",
      },
    ]);
  });

  it("handles a parameter as a number", () => {
    const setValues = jest.fn();
    const basicFormParameters = [
      {
        path: "replicas",
        value: 1,
        type: "integer",
      },
    ];
    const wrapper = mount(<DeploymentFormBody {...props} setValues={setValues} />);
    wrapper.setState({ basicFormParameters });
    wrapper.setProps({ appValues: "replicas: 1" });
    wrapper.update();

    // Fake onChange
    const input = wrapper.find(BasicDeploymentForm).find("input");
    const onChange = input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void;
    onChange({ currentTarget: { value: "2", valueAsNumber: 2, type: "number" } } as React.FormEvent<
      HTMLInputElement
    >);

    expect(wrapper.state("basicFormParameters")).toEqual([
      {
        path: "replicas",
        value: 2,
        type: "integer",
      },
    ]);
    expect(setValues).toHaveBeenCalledWith("replicas: 2\n");
  });

  it("handles a parameter as a boolean", () => {
    const setValues = jest.fn();
    const basicFormParameters = [
      {
        path: "enableMetrics",
        value: false,
        type: "boolean",
      },
    ];
    const wrapper = mount(<DeploymentFormBody {...props} setValues={setValues} />);
    wrapper.setState({ basicFormParameters });
    wrapper.setProps({ appValues: "enableMetrics: false" });
    wrapper.update();

    // Fake onChange
    const input = wrapper.find(BasicDeploymentForm).find("input");
    const onChange = input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void;
    onChange({
      currentTarget: { value: "true", checked: true, type: "checkbox" },
    } as React.FormEvent<HTMLInputElement>);

    expect(wrapper.state("basicFormParameters")).toEqual([
      {
        path: "enableMetrics",
        value: true,
        type: "boolean",
      },
    ]);
    expect(setValues).toHaveBeenCalledWith("enableMetrics: true\n");
  });
});

it("goes back when clicking in the Back button", () => {
  const goBack = jest.fn();
  const wrapper = shallow(<DeploymentFormBody {...props} goBack={goBack} />);
  const backButton = wrapper.find(".button").filterWhere(i => i.text() === "Back");
  expect(backButton).toExist();
  // Avoid empty or submit type
  expect(backButton.prop("type")).toBe("button");
  backButton.simulate("click");
  expect(goBack).toBeCalled();
});

it("restores the default chart values when clicking on the button", () => {
  const setValues = jest.fn();
  const wrapper = shallow(
    <DeploymentFormBody
      {...props}
      setValues={setValues}
      selected={{
        ...props.selected,
        values: "foo: value",
      }}
    />,
  );

  // bypass modal
  wrapper.find(ConfirmDialog).prop("onConfirm")();

  expect(setValues).toHaveBeenCalledWith("foo: value");
});

[
  {
    description: "should merge modifications from the values and the new version defaults",
    defaultValues: "foo: bar\n",
    deployedValues: "foo: bar\nmy: var\n",
    newDefaultValues: "notFoo: bar",
    result: "notFoo: bar\nmy: var\n",
  },
  {
    description: "should modify the default values",
    defaultValues: "foo: bar\n",
    deployedValues: "foo: BAR\nmy: var\n",
    newDefaultValues: "foo: bar",
    result: "foo: BAR\nmy: var\n",
  },
  {
    description: "should delete an element in the defaults",
    defaultValues: "foo: bar\n",
    deployedValues: "my: var\n",
    newDefaultValues: "foo: bar\n",
    result: "my: var\n",
  },
  {
    description: "should add an element in an array",
    defaultValues: `foo:
  - foo1:
    bar1: value1
`,
    deployedValues: `foo:
  - foo1: 
    bar1: value1
  - foo2: 
    bar2: value2
`,
    newDefaultValues: `foo:
    - foo1:
      bar1: value1
`,
    result: `foo:
  - foo1: 
    bar1: value1
  - foo2: 
    bar2: value2
`,
  },
  {
    description: "should delete an element in an array",
    defaultValues: `foo:
  - foo1:
    bar1: value1
  - foo2:
    bar2: value2
`,
    deployedValues: `foo:
  - foo2: 
    bar2: value2
`,
    newDefaultValues: `foo:
  - foo1:
    bar1: value1
  - foo2:
    bar2: value2
`,
    result: `foo:
  - foo2: 
    bar2: value2
`,
  },
].forEach(t => {
  it(t.description, () => {
    const newSelected = {
      ...defaultProps.selected,
      versions: [chartVersion],
      version: chartVersion,
      values: t.newDefaultValues,
      schema: initialSchema,
    };
    const setValues = jest.fn();
    const wrapper = shallow(
      <DeploymentFormBody
        {...props}
        deployedValues={t.deployedValues}
        setValues={setValues}
        selected={{ versions: [] }}
      />,
    );
    // Store the modifications
    wrapper.setProps({ selected: props.selected });
    expect(setValues).toHaveBeenCalledWith(t.deployedValues);

    // Apply new version
    wrapper.setProps({ selected: newSelected });
    expect(setValues).toHaveBeenCalledWith(t.result);
  });
});
