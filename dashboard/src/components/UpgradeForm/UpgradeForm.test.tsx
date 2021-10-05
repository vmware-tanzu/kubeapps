import actions from "actions";
import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import {
  AvailablePackageDetail,
  AvailablePackageReference,
  Context,
  InstalledPackageReference,
  Maintainer,
  PackageAppVersion,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import Chart from "shared/Chart";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { FetchError, IChartState } from "shared/types";
import * as url from "shared/url";
import DeploymentFormBody from "../DeploymentFormBody/DeploymentFormBody";
import UpgradeForm, { IUpgradeFormProps } from "./UpgradeForm";

const testVersion: PackageAppVersion = {
  pkgVersion: "1.2.3",
  appVersion: "4.5.6",
};

const schema = { properties: { foo: { type: "string" } } };

const availablePkgDetails = [
  {
    name: "foo",
    categories: [""],
    displayName: "foo",
    iconUrl: "https://icon.com",
    repoUrl: "https://repo.com",
    homeUrl: "https://example.com",
    sourceUrls: ["test"],
    shortDescription: "test",
    longDescription: "test",
    availablePackageRef: {
      identifier: "foo/foo",
      context: { cluster: "", namespace: "chart-namespace" } as Context,
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    },
    valuesSchema: "test",
    defaultValues: "test",
    maintainers: [{ name: "test", email: "test" }] as Maintainer[],
    readme: "test",
    version: {
      appVersion: testVersion.appVersion,
      pkgVersion: testVersion.pkgVersion,
    } as PackageAppVersion,
  },
  {
    name: "foo",
    categories: [""],
    displayName: "foo",
    iconUrl: "https://icon.com",
    repoUrl: "https://repo.com",
    homeUrl: "https://example.com",
    sourceUrls: ["test"],
    shortDescription: "test",
    longDescription: "test",
    availablePackageRef: {
      identifier: "foo/foo",
      context: { cluster: "", namespace: "chart-namespace" } as Context,
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    },
    valuesSchema: "test",
    defaultValues: "test",
    maintainers: [{ name: "test", email: "test" }] as Maintainer[],
    readme: "test",
    version: {
      appVersion: testVersion.appVersion,
      pkgVersion: testVersion.pkgVersion,
    } as PackageAppVersion,
  },
] as AvailablePackageDetail[];

const defaultProps = {
  appCurrentVersion: "1.0.0",
  appCurrentValues: "foo: bar",
  packageId: "my-chart",
  chartsIsFetching: false,
  namespace: "default",
  cluster: "default",
  releaseName: "my-release",
  repo: "my-repo",
  repoNamespace: "kubeapps",
  error: undefined,
  apps: { isFetching: false },
  charts: { isFetching: false },
  selected: {
    versions: [{ appVersion: "10.0.0", pkgVersion: "1.2.3" }],
    availablePackageDetail: { name: "test" } as AvailablePackageDetail,
  } as IChartState["selected"],
  deployed: {} as IChartState["deployed"],
  plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
} as IUpgradeFormProps;

const populatedProps = {
  ...defaultProps,
  selected: {
    error: undefined,
    availablePackageDetail: availablePkgDetails[0],
    pkgVersion: testVersion.pkgVersion,
    appVersion: testVersion.appVersion,
    readme: "readme",
    readmeError: undefined,
    values: "initial: values",
    versions: [testVersion],
    schema: schema as any,
  } as IChartState["selected"],
  deployed: {
    chartVersion: availablePkgDetails[0],
    schema: schema as any,
    values: "foo:",
  } as IChartState["deployed"],
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
          selected={{ ...defaultProps.selected, availablePackageDetail: undefined }}
        />,
      ).find(LoadingWrapper),
    ).toExist();
  });
});

it("fetches the available versions", () => {
  const getAvailablePackageVersions = jest.fn();
  Chart.getAvailablePackageVersions = getAvailablePackageVersions;
  mountWrapper(defaultStore, <UpgradeForm {...defaultProps} />);
  expect(getAvailablePackageVersions).toHaveBeenCalledWith({
    context: {
      cluster: defaultProps.cluster,
      namespace: defaultProps.repoNamespace,
    },
    identifier: defaultProps.packageId,
    plugin: defaultProps.plugin,
  } as AvailablePackageReference);
});

it("fetches the current chart version even if there is already one in the state", () => {
  const deployed = {
    chartVersion: availablePkgDetails[1],
  };

  const selected = {
    availablePackageDetail: availablePkgDetails[0],
    pkgVersion: testVersion.pkgVersion,
    appVersion: testVersion.appVersion,
    readme: "readme",
    values: "values",
    versions: [testVersion],
    schema: schema as any,
  };

  const getAvailablePackageDetail = jest.fn();
  Chart.getAvailablePackageDetail = getAvailablePackageDetail;
  mountWrapper(
    defaultStore,
    <UpgradeForm {...defaultProps} selected={selected} deployed={deployed} />,
  );
  expect(getAvailablePackageDetail).toHaveBeenCalledWith(
    {
      context: {
        cluster: availablePkgDetails[0].availablePackageRef?.context?.cluster,
        namespace: defaultProps.repoNamespace,
      },
      identifier: defaultProps.packageId,
      plugin: defaultProps.plugin,
    } as AvailablePackageReference,
    deployed.chartVersion.version?.pkgVersion,
  );
});

describe("renders an error", () => {
  it("renders an alert if the deployment failed", () => {
    const selected = {
      availablePackageDetail: { name: "foo" } as AvailablePackageDetail,
      pkgVersion: "10.0.0",
      appVersion: "1.2.3",
      versions: [{ appVersion: "10.0.0", pkgVersion: "1.2.3" }],
      schema: schema as any,
      error: new FetchError("wrong format!"),
    };
    const wrapper = mountWrapper(
      defaultStore,
      <UpgradeForm {...defaultProps} selected={selected} />,
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
  act(() => {
    const handleValuesChange: (v: string) => void = wrapper
      .find(DeploymentFormBody)
      .prop("setValues");
    handleValuesChange("foo: bar");
  });
  expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("initial: values\nfoo: bar\n");
});

it("triggers an upgrade when submitting the form", async () => {
  const mockDispatch = jest.fn().mockReturnValue(true);
  jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  const { namespace, releaseName } = defaultProps;
  const appValues = "initial: values\nfoo: bar\n";
  const upgradeApp = jest.spyOn(actions.apps, "upgradeApp").mockImplementation(() => {
    return jest.fn();
  });
  const wrapper = mountWrapper(
    defaultStore,
    <UpgradeForm {...populatedProps} namespace={namespace} />,
  );

  await act(async () => {
    const handleValuesChange: (v: string) => void = wrapper
      .find(DeploymentFormBody)
      .prop("setValues");
    handleValuesChange(appValues);
    // Simulating "submit" causes a console.warning
    (wrapper.find("form").prop("onSubmit") as (e: any) => Promise<void>)({
      preventDefault: jest.fn(),
    });
  });
  expect(upgradeApp).toHaveBeenCalledWith(
    {
      context: { cluster: defaultProps.cluster, namespace: namespace },
      identifier: releaseName,
      plugin: { name: "my.plugin", version: "0.0.1" },
    } as InstalledPackageReference,
    availablePkgDetails[0],
    "kubeapps",
    appValues,
    schema,
  );
  expect(mockDispatch).toHaveBeenCalledWith({
    payload: {
      args: [
        url.app.apps.get({
          context: { cluster: defaultProps.cluster, namespace: namespace },
          identifier: releaseName,
          plugin: { name: "my.plugin", version: "0.0.1" },
        } as InstalledPackageReference),
      ],
      method: "push",
    },
    type: "@@router/CALL_HISTORY_METHOD",
  });
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
    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toEqual(
      "initial: values\n" + currentValues,
    );
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
        selected={{
          versions: [testVersion],
          availablePackageDetail: availablePkgDetails[1],
          values: defaultValues,
        }}
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
    wrapper.setProps({ selected: { versions: [testVersion], version: availablePkgDetails[1] } });
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
      result: [`foo:`, `  - foo1: `, `    bar1: value1`, `  - foo2: `, `    bar2: value2`, ``].join(
        "\n",
      ),
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
      result: [`foo:`, `  - foo1: `, `    bar1: value1`, ``].join("\n"),
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
        version: availablePkgDetails[1],
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
