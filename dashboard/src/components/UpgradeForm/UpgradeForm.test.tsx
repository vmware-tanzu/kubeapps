import { mount, shallow } from "enzyme";
import * as React from "react";
import itBehavesLike from "../../shared/specs";

import { IChartState, IChartVersion, UnprocessableEntity } from "../../shared/types";
import * as url from "../../shared/url";
import DeploymentFormBody from "../DeploymentFormBody/DeploymentFormBody";
import { ErrorSelector } from "../ErrorAlert";
import UpgradeForm, { IUpgradeFormProps } from "./UpgradeForm";

const versions = [{ id: "foo", attributes: { version: "1.2.3" } }] as IChartVersion[];

const defaultProps = {
  appCurrentVersion: "1.0.0",
  appCurrentValues: "foo: bar",
  chartName: "my-chart",
  chartsIsFetching: false,
  namespace: "default",
  cluster: "default",
  releaseName: "my-release",
  repo: "my-repo",
  repoNamespace: "kubeapps",
  selected: { versions } as IChartState["selected"],
  deployed: {} as IChartState["deployed"],
  upgradeApp: jest.fn(),
  push: jest.fn(),
  goBack: jest.fn(),
  fetchChartVersions: jest.fn(),
  getChartVersion: jest.fn(),
  error: undefined,
} as IUpgradeFormProps;

itBehavesLike("aLoadingComponent", {
  component: UpgradeForm,
  props: { ...defaultProps, selected: { versions: [] } },
});

it("fetches the available versions", () => {
  const fetchChartVersions = jest.fn();
  shallow(<UpgradeForm {...defaultProps} fetchChartVersions={fetchChartVersions} />);
  expect(fetchChartVersions).toHaveBeenCalledWith(
    defaultProps.repoNamespace,
    `${defaultProps.repo}/${defaultProps.chartName}`,
  );
});

describe("renders an error", () => {
  it("renders a custom error if the deployment failed", () => {
    const wrapper = shallow(
      <UpgradeForm
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
      "Sorry! Something went wrong processing my-release",
    );
    expect(wrapper.find(ErrorSelector).html()).toContain("wrong format!");
  });
});

it("renders the full UpgradeForm", () => {
  const wrapper = shallow(
    <UpgradeForm {...defaultProps} selected={{ versions, version: versions[0] }} />,
  );
  expect(wrapper).toMatchSnapshot();
});

it("defaults the upgrade version to the current version", () => {
  // helm upgrade is the only way to update the values.yaml, so upgrade is
  // often used by users to update values only, so we can't default to the
  // latest version on the assumption that they always want to upgrade.
  const wrapper = shallow(
    <UpgradeForm {...defaultProps} selected={{ versions, version: versions[0] }} />,
  );

  expect(wrapper.find(DeploymentFormBody).prop("chartVersion")).toBe("1.0.0");
});

it("forwards the appValues when modified", () => {
  const wrapper = shallow(<UpgradeForm {...defaultProps} />);
  const handleValuesChange: (v: string) => void = wrapper
    .find(DeploymentFormBody)
    .prop("setValues");
  handleValuesChange("foo: bar");

  expect(wrapper.state("appValues")).toBe("foo: bar");
  expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");
});

it("triggers an upgrade when submitting the form", done => {
  const { namespace, releaseName } = defaultProps;
  const appValues = "foo: bar";
  const schema = { properties: { foo: { type: "string" } } };
  const upgradeApp = jest.fn().mockReturnValue(true);
  const push = jest.fn();
  const wrapper = mount(
    <UpgradeForm
      {...defaultProps}
      selected={{ versions, version: versions[0], schema }}
      upgradeApp={upgradeApp}
      push={push}
      namespace={namespace}
    />,
  );
  wrapper.setState({ releaseName, appValues });
  wrapper.find("form").simulate("submit");
  expect(upgradeApp).toHaveBeenCalledWith(
    defaultProps.cluster,
    namespace,
    versions[0],
    "kubeapps",
    releaseName,
    appValues,
    schema,
  );
  setTimeout(() => {
    expect(push).toHaveBeenCalledWith(
      url.app.apps.get(defaultProps.cluster, namespace, releaseName),
    );
    done();
  }, 1);
});

describe("when receiving new props", () => {
  it("should calculate the modifications from the default and the current values", () => {
    const currentValues = "a: b\nc: d\n";
    const defaultValues = "a: b\n";
    const expectedModifications = [{ op: "add", path: "/c", value: "d" }];
    const wrapper = shallow(<UpgradeForm {...defaultProps} appCurrentValues={currentValues} />);
    wrapper.setProps({ deployed: { values: defaultValues } });

    expect(wrapper.state("modifications")).toEqual(expectedModifications);
    expect(wrapper.state("appValues")).toEqual(currentValues);
  });

  it("should apply modifications if a new version is selected", () => {
    const defaultValues = "a: b\n";
    const modifications = [{ op: "add", path: "/c", value: "d" }];
    const wrapper = shallow(<UpgradeForm {...defaultProps} />);
    wrapper.setState({ modifications });
    wrapper.setProps({ selected: { versions: [], version: {}, values: defaultValues } });

    expect(wrapper.state("appValues")).toEqual("a: b\nc: d\n");
  });

  it("won't apply changes if the values have been manually modified", () => {
    const userValues = "a: b\n";
    const modifications = [{ op: "add", path: "/c", value: "d" }];
    const wrapper = shallow(<UpgradeForm {...defaultProps} />);
    wrapper.setState({ modifications, valuesModified: true, appValues: userValues });
    wrapper.setProps({ selected: { versions: [], version: {} } });

    expect(wrapper.state("appValues")).toEqual(userValues);
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
  - foo1: 
    bar1: value1
`,
      newDefaultValues: `foo:
  - foo1:
    bar1: value1
  - foo2:
    bar2: value2
`,
      result: `foo:
  - foo1: 
    bar1: value1
`,
    },
    {
      description: "set a value with dots and slashes in the key",
      defaultValues: "foo.bar/foobar: ",
      deployedValues: "foo.bar/foobar: value",
      newDefaultValues: "foo.bar/foobar: ",
      result: "foo.bar/foobar: value\n",
    },
  ].forEach(t => {
    it(t.description, () => {
      const deployed = {
        values: t.defaultValues,
        requested: true,
      };
      const newSelected = {
        ...defaultProps.selected,
        version: { trigger: "change" },
        values: t.newDefaultValues,
      };
      const wrapper = shallow(
        <UpgradeForm {...defaultProps} appCurrentValues={t.deployedValues} />,
      );
      wrapper.setProps({ deployed });

      // Apply new version
      wrapper.setProps({ selected: newSelected });
      expect(wrapper.state("appValues")).toEqual(t.result);
    });
  });
});

it("shows, by default, the default values of the deployed chart plus any modification", () => {
  const wrapper = shallow(<UpgradeForm {...defaultProps} appCurrentValues="foo: not-bar" />);
  wrapper.setProps({ deployed: { values: "# A comment\nfoo: bar\n" } as IChartState["deployed"] });
  const expectedValues = "# A comment\nfoo: not-bar\n";
  expect(wrapper.find(DeploymentFormBody).prop("deployedValues")).toBe(expectedValues);
});
