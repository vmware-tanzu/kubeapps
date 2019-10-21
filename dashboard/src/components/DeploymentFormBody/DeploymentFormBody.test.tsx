import { mount, shallow } from "enzyme";
import * as React from "react";

import { Tab, Tabs } from "react-tabs";
import itBehavesLike from "../../shared/specs";
import { IChartState, IChartVersion, NotFoundError } from "../../shared/types";
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
  originalValues: undefined,
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
const initialSchema = { properties: { foo: { type: "string", form: "foo" } } };
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
    const basicFormParameters = {
      foo: {
        form: "foo",
        path: "foo",
        value: "bar",
        type: "string",
      },
    };
    expect(localState.basicFormParameters).toEqual(basicFormParameters);
  });

  describe("when the user has not modified any value", () => {
    it("selects the original values if the version doesn't change", () => {
      const setValues = jest.fn();
      const originalValues = "foo: notBar";
      const wrapper = shallow(
        <DeploymentFormBody {...props} setValues={setValues} originalValues={originalValues} />,
      );
      wrapper.setProps({
        selected: {
          ...props.selected,
          values: "foo: ignored-value",
        },
      });
      const basicFormParameters = {
        foo: {
          form: "foo",
          path: "foo",
          value: "notBar",
          type: "string",
        },
      };
      const localState: IDeploymentFormBodyState = wrapper.instance()
        .state as IDeploymentFormBodyState;
      expect(localState.basicFormParameters).toEqual(basicFormParameters);
      expect(setValues).toHaveBeenCalledWith("foo: notBar");
    });

    it("uses the chart default values when original values are not defined", () => {
      const setValues = jest.fn();
      const wrapper = shallow(
        <DeploymentFormBody {...props} setValues={setValues} originalValues={undefined} />,
      );
      wrapper.setProps({
        selected: {
          ...props.selected,
          values: "foo: notBar",
        },
      });
      const basicFormParameters = {
        foo: {
          form: "foo",
          path: "foo",
          value: "notBar",
          type: "string",
        },
      };
      const localState: IDeploymentFormBodyState = wrapper.instance()
        .state as IDeploymentFormBodyState;
      expect(localState.basicFormParameters).toEqual(basicFormParameters);
      expect(setValues).toHaveBeenCalledWith("foo: notBar");
    });
  });

  describe("when the user has modified the values", () => {
    it("will ignore original or default values", () => {
      const setValues = jest.fn();
      const originalValues = "foo: ignored-value";
      const modifiedValues = "foo: notBar";
      const wrapper = shallow(
        <DeploymentFormBody
          {...props}
          setValues={setValues}
          originalValues={originalValues}
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
      const basicFormParameters = {
        foo: {
          form: "foo",
          path: "foo",
          value: "notBar",
          type: "string",
        },
      };
      const localState: IDeploymentFormBodyState = wrapper.instance()
        .state as IDeploymentFormBodyState;
      expect(localState.basicFormParameters).toEqual(basicFormParameters);
      expect(setValues).not.toHaveBeenCalled();
    });
  });
});

describe("when the basic form is enabled", () => {
  it("renders the different tabs", () => {
    const basicFormParameters = {
      username: {
        path: "wordpressUsername",
        value: "user",
      },
    };
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
    const basicFormParameters = {
      username: {
        path: "wordpressUsername",
        value: "user",
      },
    };
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

    expect(wrapper.state("basicFormParameters")).toEqual({
      username: {
        path: "wordpressUsername",
        value: "foo",
      },
    });
    expect(setValuesModified).toHaveBeenCalled();
    expect(setValues).toHaveBeenCalledWith("wordpressUsername: foo\n");
  });

  it("should update existing params if the app values change and the user clicks on the Basic tab", () => {
    const testProps = {
      ...props,
      selected: {
        ...props.selected,
        schema: { properties: { wordpressUsername: { type: "string", form: "username" } } },
      },
    };
    const basicFormParameters = {
      username: {
        path: "wordpressUsername",
        value: "user",
      },
    };
    const wrapper = mount(<DeploymentFormBody {...testProps} />);
    wrapper.setState({ basicFormParameters });
    wrapper.setProps({ appValues: "wordpressUsername: foo" });
    wrapper.update();

    const tab = wrapper
      .find(Tab)
      .findWhere(t => !!t.text().match("Basic"))
      .first();
    tab.simulate("click");

    expect(wrapper.state("basicFormParameters")).toMatchObject({
      username: {
        path: "wordpressUsername",
        value: "foo",
      },
    });
  });

  it("handles a parameter as a number", () => {
    const setValues = jest.fn();
    const basicFormParameters = {
      replicas: {
        path: "replicas",
        value: 1,
        type: "integer",
      },
    };
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

    expect(wrapper.state("basicFormParameters")).toEqual({
      replicas: {
        path: "replicas",
        value: 2,
        type: "integer",
      },
    });
    expect(setValues).toHaveBeenCalledWith("replicas: 2\n");
  });

  it("handles a parameter as a boolean", () => {
    const setValues = jest.fn();
    const basicFormParameters = {
      enableMetrics: {
        path: "enableMetrics",
        value: false,
        type: "boolean",
      },
    };
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

    expect(wrapper.state("basicFormParameters")).toEqual({
      enableMetrics: {
        path: "enableMetrics",
        value: true,
        type: "boolean",
      },
    });
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
