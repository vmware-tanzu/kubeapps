import { mount, shallow } from "enzyme";
import * as React from "react";
import itBehavesLike from "../../shared/specs";

import Alert from "components/js/Alert";
import { act } from "react-dom/test-utils";
import { IChartState, IChartVersion } from "../../shared/types";
import * as url from "../../shared/url";
import DeploymentFormBody from "../DeploymentFormBody/DeploymentFormBody.v2";
import UpgradeForm, { IUpgradeFormProps } from "./UpgradeForm.v2";

const versions = [
  {
    id: "foo",
    attributes: { version: "1.2.3" },
    relationships: { chart: { data: { repo: { name: "bitnami" } } } },
  },
  {
    id: "foo",
    attributes: { version: "1.2.4" },
    relationships: { chart: { data: { repo: { name: "bitnami" } } } },
  },
] as IChartVersion[];

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

const schema = { properties: { foo: { type: "string" } } };

const populatedProps = {
  ...defaultProps,
  selected: { versions, version: versions[0], schema },
};

itBehavesLike("aLoadingComponent", {
  component: UpgradeForm,
  props: { ...defaultProps, selected: { versions: [] } },
});

it("fetches the available versions", () => {
  const fetchChartVersions = jest.fn();
  mount(<UpgradeForm {...defaultProps} fetchChartVersions={fetchChartVersions} />);
  expect(fetchChartVersions).toHaveBeenCalledWith(
    defaultProps.repoNamespace,
    `${defaultProps.repo}/${defaultProps.chartName}`,
  );
});

describe("renders an error", () => {
  it("renders an alert if the deployment failed", () => {
    const wrapper = shallow(
      <UpgradeForm
        {...defaultProps}
        selected={
          {
            version: { attributes: {} },
            versions: [{ id: "foo", attributes: {} }],
          } as IChartState["selected"]
        }
        error={new Error("wrong format!")}
      />,
    );
    expect(wrapper.find(Alert).exists()).toBe(true);
    expect(wrapper.find(Alert).html()).toContain("wrong format!");
  });
});

it("defaults the upgrade version to the current version", () => {
  // helm upgrade is the only way to update the values.yaml, so upgrade is
  // often used by users to update values only, so we can't default to the
  // latest version on the assumption that they always want to upgrade.
  const wrapper = shallow(<UpgradeForm {...populatedProps} />);

  expect(wrapper.find(DeploymentFormBody).prop("chartVersion")).toBe("1.0.0");
});

it("forwards the appValues when modified", () => {
  const wrapper = shallow(<UpgradeForm {...populatedProps} />);
  const handleValuesChange: (v: string) => void = wrapper
    .find(DeploymentFormBody)
    .prop("setValues");
  handleValuesChange("foo: bar");

  expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");
});

it("triggers an upgrade when submitting the form", async () => {
  const { namespace, releaseName } = defaultProps;
  const appValues = "foo: bar";
  const upgradeApp = jest.fn().mockReturnValue(true);
  const push = jest.fn();
  const wrapper = mount(
    <UpgradeForm {...populatedProps} upgradeApp={upgradeApp} push={push} namespace={namespace} />,
  );
  const handleValuesChange: (v: string) => void = wrapper
    .find(DeploymentFormBody)
    .prop("setValues");
  handleValuesChange(appValues);

  await act(async () => {
    // Simulating "submit" causes a console.warning
    await (wrapper.find("form").prop("onSubmit") as (e: any) => Promise<void>)({
      preventDefault: jest.fn(),
    });
  });
  expect(upgradeApp).toHaveBeenCalledWith(
    defaultProps.cluster,
    namespace,
    versions[0],
    "kubeapps",
    releaseName,
    appValues,
    schema,
  );
  expect(push).toHaveBeenCalledWith(url.app.apps.get(defaultProps.cluster, namespace, releaseName));
});

describe("when receiving new props", () => {
  it("should calculate the modifications from the default and the current values", () => {
    const currentValues = "a: b\nc: d\n";
    const defaultValues = "a: b\n";
    const wrapper = mount(<UpgradeForm {...populatedProps} appCurrentValues={currentValues} />);
    wrapper.setProps({ deployed: { values: defaultValues } });

    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toEqual(currentValues);
  });

  it("should apply modifications if a new version is selected", () => {
    const defaultValues = "a: b\n";
    const deployedValues = "a: B\n";
    const currentValues = "a: B\nc: d\n";
    const wrapper = mount(
      <UpgradeForm
        {...populatedProps}
        deployed={{ values: deployedValues }}
        appCurrentValues={currentValues}
      />,
    );
    wrapper.setProps({ selected: { versions, version: versions[1], values: defaultValues } });
    wrapper.update();
    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toEqual("a: b\nc: d\n");
  });

  it("won't apply changes if the values have been manually modified", () => {
    const userValues = "a: b\n";
    const wrapper = mount(<UpgradeForm {...populatedProps} />);
    act(() => {
      const handleValuesChange: (v: string) => void = wrapper
        .find(DeploymentFormBody)
        .prop("setValues");
      handleValuesChange(userValues);
      const setValuesModified: () => void = wrapper
        .find(DeploymentFormBody)
        .prop("setValuesModified");
      setValuesModified();
    });
    wrapper.setProps({ selected: { versions, version: versions[1] } });
    wrapper.update();

    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toEqual(userValues);
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
        ...populatedProps.selected,
        version: versions[1],
        values: t.newDefaultValues,
      };
      const wrapper = mount(
        <UpgradeForm {...populatedProps} appCurrentValues={t.deployedValues} />,
      );
      wrapper.setProps({ deployed });

      // Apply new version
      wrapper.setProps({ selected: newSelected });
      wrapper.update();
      expect(wrapper.find(DeploymentFormBody).prop("appValues")).toEqual(t.result);
    });
  });
});

it("shows, by default, the default values of the deployed chart plus any modification", () => {
  const wrapper = mount(
    <UpgradeForm
      {...populatedProps}
      deployed={{ values: "# A comment\nfoo: bar\n" }}
      appCurrentValues="foo: not-bar"
    />,
  );
  const expectedValues = "# A comment\nfoo: not-bar\n";
  expect(wrapper.find(DeploymentFormBody).prop("deployedValues")).toBe(expectedValues);
});
