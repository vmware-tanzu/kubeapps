import actions from "actions";
import ChartHeader from "components/ChartView/ChartHeader";
import Alert from "components/js/Alert";
import {
  AvailablePackageDetail,
  PackageAppVersion,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { createMemoryHistory } from "history";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import * as ReactRouter from "react-router";
import { MemoryRouter, Route, Router } from "react-router";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import { FetchError } from "shared/types";
import * as url from "shared/url";
import DeploymentFormBody from "../DeploymentFormBody/DeploymentFormBody";
import DeploymentForm from "./DeploymentForm";

const defaultProps = {
  pkgName: "foo",
  cluster: "default",
  namespace: "default",
  repo: "repo",
  releaseName: "my-release",
};
const routePathParam = `/c/${defaultProps.cluster}/ns/${defaultProps.namespace}/apps/new/${defaultProps.repo}/${defaultProps.pkgName}/versions/`;
const routePath = "/c/:cluster/ns/:namespace/apps/new/:repo/:id/versions";
const history = createMemoryHistory({ initialEntries: [routePathParam] });

let spyOnUseDispatch: jest.SpyInstance;
let spyOnUseHistory: jest.SpyInstance;

beforeEach(() => {
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  spyOnUseHistory = jest
    .spyOn(ReactRouter, "useHistory")
    .mockReturnValue({ push: jest.fn() } as any);
});

afterEach(() => {
  jest.resetAllMocks();
  spyOnUseDispatch.mockRestore();
  spyOnUseHistory.mockRestore();
});

it("fetches the available versions", () => {
  const fetchChartVersions = jest.fn();
  actions.charts.fetchChartVersion = fetchChartVersions;

  mountWrapper(
    getStore({}),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <DeploymentForm />
      </Route>
    </MemoryRouter>,
  );

  expect(fetchChartVersions).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.namespace,
    `${defaultProps.repo}/${defaultProps.pkgName}`,
    undefined,
  );
});

describe("renders an error", () => {
  it("renders a custom error if the deployment failed", () => {
    const wrapper = mountWrapper(
      getStore({ apps: { error: new FetchError("wrong format!") } }),
      <DeploymentForm />,
    );
    expect(wrapper.find(Alert).exists()).toBe(true);
    expect(wrapper.find(Alert).html()).toContain("wrong format!");
  });

  it("renders a fetch error only", () => {
    const wrapper = mountWrapper(
      getStore({ apps: { error: new FetchError("wrong format!") } }),
      <DeploymentForm />,
    );
    expect(wrapper.find(Alert)).toExist();
    expect(wrapper.find(ChartHeader)).not.toExist();
  });

  it("forwards the appValues when modified", () => {
    const selected = {
      versions: [{ appVersion: "10.0.0", pkgVersion: "1.2.3" } as PackageAppVersion],
      availablePackageDetail: { name: "test" } as AvailablePackageDetail,
      pkgVersion: "1.2.4",
      values: "bar: foo",
    };
    const wrapper = mountWrapper(getStore({ charts: { selected: selected } }), <DeploymentForm />);

    const handleValuesChange: (v: string) => void = wrapper
      .find(DeploymentFormBody)
      .prop("setValues");
    act(() => {
      handleValuesChange("foo: bar");
    });
    wrapper.update();

    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");
  });

  it("changes values if the version changes and it has not been modified", () => {
    const selected = {
      versions: [{ appVersion: "10.0.0", pkgVersion: "1.2.3" } as PackageAppVersion],
      availablePackageDetail: { name: "test" } as AvailablePackageDetail,
      pkgVersion: "1.2.4",
      values: "bar: foo",
    };
    const wrapper = mountWrapper(getStore({ charts: { selected: selected } }), <DeploymentForm />);
    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("bar: foo");
  });

  it("keep values if the version changes", () => {
    const selected = {
      versions: [{ appVersion: "10.0.0", pkgVersion: "1.2.3" } as PackageAppVersion],
      availablePackageDetail: { name: "test" } as AvailablePackageDetail,
      pkgVersion: "1.2.4",
      values: "bar: foo",
    };
    const wrapper = mountWrapper(getStore({ charts: { selected: selected } }), <DeploymentForm />);

    const handleValuesChange: (v: string) => void = wrapper
      .find(DeploymentFormBody)
      .prop("setValues");
    const setValuesModified: () => void = wrapper
      .find(DeploymentFormBody)
      .prop("setValuesModified");
    act(() => {
      handleValuesChange("foo: bar");
      setValuesModified();
    });
    wrapper.update();
    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");

    wrapper.find("select").simulate("change", { target: { value: "1.2.4" } });
    wrapper.setProps({ selected: { ...selected, values: "bar: foo" } });
    wrapper.update();
    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");
  });

  it("triggers a deployment when submitting the form", async () => {
    const deployChart = jest.fn().mockReturnValue(true);
    const push = jest.fn();
    actions.apps.deployChart = deployChart;
    spyOnUseHistory = jest.spyOn(ReactRouter, "useHistory").mockReturnValue({ push } as any);

    const appValues = "foo: bar";
    const schema = { properties: { foo: { type: "string", form: true } } };
    const availablePackageDetail = { name: "test" };
    const selected = {
      versions: [{ appVersion: "10.0.0", pkgVersion: "1.2.3" } as PackageAppVersion],
      availablePackageDetail: availablePackageDetail,
      pkgVersion: "1.2.4",
      values: appValues,
      schema: schema,
    };

    const wrapper = mountWrapper(
      getStore({ charts: { selected: selected } }),

      <Router history={history}>
        <Route path="/c/:cluster/ns/:namespace/apps/new/:repo/:id/versions">
          <DeploymentForm />
        </Route>
      </Router>,
    );

    const handleValuesChange: (v: string) => void = wrapper
      .find(DeploymentFormBody)
      .prop("setValues");
    act(() => {
      handleValuesChange("foo: bar");
    });

    wrapper
      .find("#releaseName")
      .simulate("change", { target: { value: defaultProps.releaseName } });

    wrapper.update();

    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");
    expect(wrapper.find(DeploymentForm).find("#releaseName").prop("value")).toBe(
      defaultProps.releaseName,
    );

    await act(async () => {
      // Simulating "submit" causes a console.warning
      await (wrapper.find("form").prop("onSubmit") as (e: any) => Promise<void>)({
        preventDefault: jest.fn(),
      });
    });

    expect(deployChart).toHaveBeenCalledWith(
      defaultProps.cluster,
      defaultProps.namespace,
      availablePackageDetail,
      defaultProps.releaseName,
      appValues,
      schema,
    );

    expect(history.location.pathname).toBe("/c/default/ns/default/apps/new/repo/foo/versions/");
  });
});
