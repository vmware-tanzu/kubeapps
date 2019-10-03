import { mount, shallow } from "enzyme";
import * as Moniker from "moniker-native";
import * as React from "react";

import { Tabs } from "react-tabs";
import itBehavesLike from "../../shared/specs";
import { IChartState, IChartVersion, NotFoundError, UnprocessableEntity } from "../../shared/types";
import { ErrorSelector } from "../ErrorAlert";
import ErrorPageHeader from "../ErrorAlert/ErrorAlertHeader";
import LoadingWrapper from "../LoadingWrapper";
import AdvancedDeploymentForm from "./AdvancedDeploymentForm";
import BasicDeploymentForm from "./BasicDeploymentForm";
import DeploymentForm, { IDeploymentFormProps, IDeploymentFormState } from "./DeploymentForm";

const defaultProps = {
  kubeappsNamespace: "kubeapps",
  chartID: "foo",
  chartVersion: "1.0.0",
  error: undefined,
  selected: {} as IChartState["selected"],
  deployChart: jest.fn(),
  push: jest.fn(),
  fetchChartVersions: jest.fn(),
  getChartVersion: jest.fn(),
  getChartValues: jest.fn(),
  getChartSchema: jest.fn(),
  namespace: "default",
  enableBasicForm: false,
};
const versions = [{ id: "foo", attributes: { version: "1.2.3" } }] as IChartVersion[];
let monikerChooseMock: jest.Mock;

itBehavesLike("aLoadingComponent", { component: DeploymentForm, props: defaultProps });

beforeEach(() => {
  monikerChooseMock = jest.fn();
  Moniker.choose = monikerChooseMock;
});

afterEach(() => {
  jest.resetAllMocks();
});

describe("renders an error", () => {
  it("renders an error if it cannot find the given chart", () => {
    const wrapper = mount(
      <DeploymentForm
        {...defaultProps}
        selected={{ error: new NotFoundError() } as IChartState["selected"]}
      />,
    );
    expect(wrapper.find(ErrorPageHeader).exists()).toBe(true);
    expect(wrapper.find(ErrorPageHeader).text()).toContain('Chart "foo" (1.0.0) not found');
  });

  it("renders a generic error", () => {
    const wrapper = shallow(
      <DeploymentForm
        {...defaultProps}
        selected={{ error: new Error() } as IChartState["selected"]}
      />,
    );
    expect(wrapper.find(ErrorSelector).exists()).toBe(true);
    expect(wrapper.find(ErrorSelector).html()).toContain("Sorry! Something went wrong");
  });

  it("renders a custom error if the deployment failed", () => {
    const wrapper = shallow(
      <DeploymentForm
        {...defaultProps}
        selected={
          {
            version: { attributes: {} },
            versions: [{ id: "foo", attributes: {} }],
          } as IChartState["selected"]
        }
        error={new UnprocessableEntity("wrong format!")}
      />,
    );
    wrapper.setState({ latestSubmittedReleaseName: "my-app" });
    expect(wrapper.find(ErrorSelector).exists()).toBe(true);
    expect(wrapper.find(ErrorSelector).html()).toContain(
      "Sorry! Something went wrong processing my-app",
    );
    expect(wrapper.find(ErrorSelector).html()).toContain("wrong format!");
  });

  it("the error does not change if the release name changes", () => {
    const expectedErrorMsg = "Sorry! Something went wrong processing my-app";

    const wrapper = shallow(
      <DeploymentForm
        {...defaultProps}
        selected={
          {
            version: { attributes: {} },
            versions: [{ id: "foo", attributes: {} }],
          } as IChartState["selected"]
        }
        error={new UnprocessableEntity("wrong format!")}
      />,
    );

    wrapper.setState({ latestSubmittedReleaseName: "my-app" });
    expect(wrapper.find(ErrorSelector).exists()).toBe(true);
    expect(wrapper.find(ErrorSelector).html()).toContain(expectedErrorMsg);
    wrapper.setState({ releaseName: "another-app" });
    expect(wrapper.find(ErrorSelector).html()).toContain(expectedErrorMsg);
  });
});

it("renders the full DeploymentForm", () => {
  const wrapper = shallow(
    <DeploymentForm {...defaultProps} selected={{ versions, version: versions[0] }} />,
  );
  expect(wrapper).toMatchSnapshot();
});

it("renders a release name by default, relying in Monickers output", () => {
  monikerChooseMock.mockImplementationOnce(() => "foo").mockImplementationOnce(() => "bar");

  let wrapper = shallow(
    <DeploymentForm {...defaultProps} selected={{ versions, version: versions[0] }} />,
  );
  const name1 = wrapper.state("releaseName") as string;
  expect(name1).toBe("foo");

  // When reloading the name should change
  wrapper = shallow(
    <DeploymentForm {...defaultProps} selected={{ versions, version: versions[0] }} />,
  );
  const name2 = wrapper.state("releaseName") as string;
  expect(name2).toBe("bar");
});

const initialValues = "some yaml text";
const chartVersion = {
  id: "foo",
  attributes: { version: "1.0", app_version: "1.0", created: "1" },
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
const props: IDeploymentFormProps = {
  ...defaultProps,
  selected: {
    ...defaultProps.selected,
    versions: [chartVersion],
    version: chartVersion,
    values: initialValues,
  },
};
describe("stores modified values locally", () => {
  it("initializes the local values from props when props set", () => {
    const wrapper = shallow(<DeploymentForm {...props} />);

    wrapper.setProps(props);

    const localState: IDeploymentFormState = wrapper.instance().state as IDeploymentFormState;
    expect(localState.appValues).toEqual(initialValues);
  });

  it("updates initial values from props if not modified", () => {
    const wrapper = shallow(<DeploymentForm {...props} />);

    const updatedValuesFromProps = "some other yaml";
    wrapper.setProps({
      ...props,
      selected: {
        ...props.selected,
        values: updatedValuesFromProps,
      },
    });

    const localState: IDeploymentFormState = wrapper.instance().state as IDeploymentFormState;
    expect(localState.appValues).toEqual(updatedValuesFromProps);
  });

  it("does not update values from props if they have been modified in local state", () => {
    const wrapper = shallow(<DeploymentForm {...props} />);
    const modifiedValues = "user-modified values.yaml";
    const form: DeploymentForm = wrapper.instance() as DeploymentForm;
    form.handleValuesChange(modifiedValues);

    const updatedValuesFromProps = "some other yaml";
    wrapper.setProps({
      ...props,
      selected: {
        ...props.selected,
        values: updatedValuesFromProps,
      },
    });

    const localState: IDeploymentFormState = wrapper.instance().state as IDeploymentFormState;
    expect(localState.appValues).not.toEqual(updatedValuesFromProps);
    expect(localState.appValues).toEqual(modifiedValues);
  });
});

describe("when the basic form is not enabled", () => {
  it("the advanced editor should be shown", () => {
    const wrapper = shallow(<DeploymentForm {...props} enableBasicForm={false} />);
    expect(wrapper.find(LoadingWrapper)).not.toExist();
    expect(wrapper.find(AdvancedDeploymentForm)).toExist();
  });

  it("should not show the basic/advanced tabs", () => {
    const wrapper = shallow(<DeploymentForm {...props} enableBasicForm={false} />);
    expect(wrapper.find(LoadingWrapper)).not.toExist();
    expect(wrapper.find(Tabs)).not.toExist();
  });
});

describe("when the basic form is enabled", () => {
  it("renders the different tabs", () => {
    const wrapper = shallow(<DeploymentForm {...props} enableBasicForm={true} />);
    expect(wrapper.find(LoadingWrapper)).not.toExist();
    expect(wrapper.find(Tabs)).toExist();
  });

  it("changes the parameter value", () => {
    const basicFormParameters = {
      username: {
        path: "wordpressUsername",
        value: "user",
      },
    };
    const wrapper = mount(<DeploymentForm {...props} enableBasicForm={true} />);
    wrapper.setState({ appValues: "wordpressUsername: user", basicFormParameters });
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
    expect(wrapper.state("appValues")).toBe("wordpressUsername: foo\n");
  });
});
