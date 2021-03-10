import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import { act } from "react-dom/test-utils";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IChartState, IChartVersion } from "../../shared/types";
import * as url from "../../shared/url";
import DeploymentFormBody from "../DeploymentFormBody/DeploymentFormBody";
import UpgradeForm, { IUpgradeFormProps } from "./UpgradeForm";

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

describe("it behaves like a loading component", () => {
  it("if the app is being fetched", () => {
    expect(
      mountWrapper(
        getStore({ apps: { isFetching: true } }),
        <UpgradeForm {...defaultProps} />,
      ).find(LoadingWrapper),
    ).toExist();
  });

  it("if the chart is being fetched", () => {
    expect(
      mountWrapper(
        getStore({ charts: { isFetching: true } }),
        <UpgradeForm {...defaultProps} />,
      ).find(LoadingWrapper),
    ).toExist();
  });

  it("if there are no versions", () => {
    expect(
      mountWrapper(
        defaultStore,
        <UpgradeForm {...defaultProps} selected={{ ...defaultProps.selected, versions: [] }} />,
      ).find(LoadingWrapper),
    ).toExist();
  });

  it("if there is no version", () => {
    expect(
      mountWrapper(
        defaultStore,
        <UpgradeForm
          {...defaultProps}
          selected={{ ...defaultProps.selected, version: undefined }}
        />,
      ).find(LoadingWrapper),
    ).toExist();
  });
});

it("fetches the available versions", () => {
  const fetchChartVersions = jest.fn();
  mountWrapper(
    defaultStore,
    <UpgradeForm {...defaultProps} fetchChartVersions={fetchChartVersions} />,
  );
  expect(fetchChartVersions).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.repoNamespace,
    `${defaultProps.repo}/${defaultProps.chartName}`,
  );
});

it("fetches the current chart version even if there is already one in the state", () => {
  const deployed = {
    chartVersion: {
      attributes: {
        version: "1.0.0",
      },
    },
  } as any;
  const selected = {
    version: { attributes: {}, relationships: { chart: { data: { repo: { name: "" } } } } },
    versions: [{ id: "foo", attributes: {} }],
  } as IChartState["selected"];

  const getChartVersion = jest.fn();
  mountWrapper(
    defaultStore,
    <UpgradeForm
      {...defaultProps}
      selected={selected}
      deployed={deployed}
      getChartVersion={getChartVersion}
    />,
  );
  expect(getChartVersion).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.repoNamespace,
    `${defaultProps.repo}/${defaultProps.chartName}`,
    deployed.chartVersion.attributes.version,
  );
});

describe("renders an error", () => {
  it("renders an alert if the deployment failed", () => {
    const wrapper = mountWrapper(
      defaultStore,
      <UpgradeForm
        {...defaultProps}
        selected={
          {
            version: { attributes: {}, relationships: { chart: { data: { repo: { name: "" } } } } },
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
  const wrapper = mountWrapper(defaultStore, <UpgradeForm {...populatedProps} />);

  expect(wrapper.find(DeploymentFormBody).prop("chartVersion")).toBe("1.0.0");
});

it("forwards the appValues when modified", () => {
  const wrapper = mountWrapper(defaultStore, <UpgradeForm {...populatedProps} />);
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
  const wrapper = mountWrapper(
    defaultStore,
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
    const wrapper = mountWrapper(
      defaultStore,
      <UpgradeForm {...populatedProps} appCurrentValues={currentValues} />,
    );
    wrapper.setProps({ deployed: { values: defaultValues } });

    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toEqual(currentValues);
  });

  it("should apply modifications if a new version is selected", () => {
    const defaultValues = "a: b\n";
    const deployedValues = "a: B\n";
    const currentValues = "a: B\nc: d\n";
    const wrapper = mountWrapper(
      defaultStore,
      <UpgradeForm
        {...populatedProps}
        deployed={{ values: deployedValues }}
        appCurrentValues={currentValues}
        selected={{ versions, version: versions[1], values: defaultValues }}
      />,
    );
    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toEqual("a: b\nc: d\n");
  });

  it("won't apply changes if the values have been manually modified", () => {
    const userValues = "a: b\n";
    const wrapper = mountWrapper(defaultStore, <UpgradeForm {...populatedProps} />);
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
      const wrapper = mountWrapper(
        defaultStore,
        <UpgradeForm
          {...populatedProps}
          appCurrentValues={t.deployedValues}
          deployed={deployed}
          selected={newSelected}
        />,
      );
      expect(wrapper.find(DeploymentFormBody).prop("appValues")).toEqual(t.result);
    });
  });
});

it("shows, by default, the default values of the deployed chart plus any modification", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <UpgradeForm
      {...populatedProps}
      deployed={{ values: "# A comment\nfoo: bar\n" }}
      appCurrentValues="foo: not-bar"
    />,
  );
  const expectedValues = "# A comment\nfoo: not-bar\n";
  expect(wrapper.find(DeploymentFormBody).prop("deployedValues")).toBe(expectedValues);
});
