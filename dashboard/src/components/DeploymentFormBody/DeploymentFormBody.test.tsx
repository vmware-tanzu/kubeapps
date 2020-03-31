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
import Differential from "./Differential";

const defaultProps = {
  chartNamespace: "chart-namespace",
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

it("marks the current version", () => {
  const wrapper = shallow(
    <DeploymentFormBody
      {...defaultProps}
      releaseVersion={versions[0].attributes.version}
      selected={{ versions, version: versions[0] }}
    />,
  );
  expect(wrapper.find("select").text()).toMatch("1.2.3 (current)");
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
          namespace: "repo-namespace",
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
      .findWhere(t => !!t.text().match("Form"))
      .first();
    tab.simulate("click");

    expect(wrapper.state("basicFormParameters")).toMatchObject([
      {
        path: "wordpressUsername",
        value: "foo",
      },
    ]);
  });

  it("should update existing params when receiving new values", () => {
    const testProps = {
      ...props,
      selected: {
        ...props.selected,
        schema: { properties: { wordpressUsername: { type: "string", form: true } } },
      },
      appValues: "",
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
    expect(wrapper.state("basicFormParameters")).toMatchObject([
      {
        path: "wordpressUsername",
        value: "foo",
      },
    ]);
  });

  it("should update existing params when receiving new values (for a new version)", () => {
    const testProps = {
      ...props,
      selected: {
        ...props.selected,
        schema: { properties: { wordpressUsername: { type: "string", form: true } } },
        values: "wordpressUsername: foo",
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
    const updatedProps = {
      selected: {
        ...props.selected,
        schema: { properties: { wordpressUsername: { type: "string", form: true } } },
        values: "wordpressUsername: bar",
      },
      appValues: "wordpressUsername: bar",
    };

    wrapper.setProps(updatedProps);
    wrapper.update();
    expect(wrapper.state("basicFormParameters")).toMatchObject([
      {
        path: "wordpressUsername",
        value: "bar",
      },
    ]);
  });

  it("should not re-render the basic params if appValues changes (because it's handled by the parameter itself)", () => {
    const testProps = {
      ...props,
      selected: {
        ...props.selected,
        schema: { properties: { wordpressUsername: { type: "string", form: true } } },
      },
      appValues: "wordpressUsername: foo",
    };
    const basicFormParameters = [
      {
        path: "wordpressUsername",
        value: "foo",
      },
    ];
    const wrapper = mount(<DeploymentFormBody {...testProps} />);
    wrapper.setState({ basicFormParameters });
    wrapper.setProps({ appValues: "wordpressUsername: bar" });
    wrapper.update();
    expect(wrapper.state("basicFormParameters")).toMatchObject([
      {
        path: "wordpressUsername",
        value: "foo",
      },
    ]);
  });

  it("handles a parameter as a number", () => {
    const setValues = jest.fn();
    const testProps = {
      ...props,
      selected: {
        ...props.selected,
        schema: { properties: { replicas: { type: "integer", form: true } } },
      },
    };
    const basicFormParameters = [
      {
        form: true,
        path: "replicas",
        value: 1,
        type: "integer",
      },
    ];
    const wrapper = mount(<DeploymentFormBody {...testProps} setValues={setValues} />);
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
        form: true,
        path: "replicas",
        value: 2,
        type: "integer",
      },
    ]);
    expect(setValues).toHaveBeenCalledWith("replicas: 2\n");
  });

  it("handles a parameter as a boolean", () => {
    const testProps = {
      ...props,
      selected: {
        ...props.selected,
        schema: { properties: { enableMetrics: { type: "boolean", form: true } } },
      },
    };
    const setValues = jest.fn();
    const basicFormParameters = [
      {
        form: true,
        path: "enableMetrics",
        value: false,
        type: "boolean",
      },
    ];
    const wrapper = mount(<DeploymentFormBody {...testProps} setValues={setValues} />);
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
        form: true,
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

describe("Changes tab", () => {
  it("should show the differences between the default chart values when deploying", () => {
    const selected = {
      ...defaultProps.selected,
      versions: [chartVersion],
      version: chartVersion,
      values: "foo: bar",
      schema: initialSchema,
    };
    const appValues = "bar: foo";
    const wrapper = shallow(
      <DeploymentFormBody {...props} selected={selected} appValues={appValues} />,
    );

    const Diff = wrapper.find(Differential);
    expect(Diff.props()).toMatchObject({
      emptyDiffText: "No changes detected from chart defaults.",
      newValues: "bar: foo",
      oldValues: "foo: bar",
      title: "Difference from chart defaults",
    });
  });

  it("should show the differences between the current release and the new one when upgrading", () => {
    const selected = {
      ...defaultProps.selected,
      versions: [chartVersion],
      version: chartVersion,
      values: "foo: bar",
      schema: initialSchema,
    };
    const deployedValues = "a: b";
    const appValues = "bar: foo";
    const wrapper = shallow(
      <DeploymentFormBody
        {...props}
        selected={selected}
        appValues={appValues}
        deployedValues={deployedValues}
      />,
    );

    const Diff = wrapper.find(Differential);
    expect(Diff.props()).toMatchObject({
      emptyDiffText: "The values for the new release are identical to the deployed version.",
      newValues: "bar: foo",
      oldValues: "a: b",
      title: "Difference from deployed version",
    });
  });
});
