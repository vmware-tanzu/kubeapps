import actions from "actions";
import Alert from "components/js/Alert";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import SelectRepoForm from "./SelectRepoForm";

const defaultProps = {
  cluster: "default",
  namespace: "default",
  chartName: "test",
};

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.operators };
beforeEach(() => {
  actions.repos = {
    ...actions.repos,
    fetchRepos: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.operators = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("should fetch only the global repository", () => {
  const fetch = jest.fn();
  actions.repos = { ...actions.repos, fetchRepos: fetch };
  const props = {
    cluster: defaultProps.cluster,
    namespace: initialState.config.kubeappsNamespace, // global
    chartName: defaultProps.chartName,
  };
  mountWrapper(defaultStore, <SelectRepoForm {...props} />);
  expect(fetch).toHaveBeenCalledWith(initialState.config.kubeappsNamespace);
});

it("should fetch repositories", () => {
  const fetch = jest.fn();
  actions.repos = { ...actions.repos, fetchRepos: fetch };
  mountWrapper(defaultStore, <SelectRepoForm {...defaultProps} />);
  expect(fetch).toHaveBeenCalledWith(defaultProps.namespace, true);
});

it("should render a loading page if fetching", () => {
  expect(
    mountWrapper(
      getStore({ repos: { isFetching: true } }),
      <SelectRepoForm {...defaultProps} />,
    ).find("LoadingWrapper"),
  ).toExist();
});

it("render an error if failed to request repos", () => {
  const wrapper = mountWrapper(
    getStore({ repos: { errors: { fetch: new Error("boom") } } }),
    <SelectRepoForm {...defaultProps} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("boom");
});

it("render a warning if there are no repos", () => {
  const wrapper = mountWrapper(defaultStore, <SelectRepoForm {...defaultProps} />);
  expect(wrapper.find(Alert)).toIncludeText("Chart repositories not found");
});

it("should select a repo", () => {
  const checkChart = jest.fn();
  actions.repos = { ...actions.repos, checkChart };
  const repo = {
    metadata: {
      name: "bitnami",
      namespace: "default",
    },
    spec: {
      url: "http://repo",
    },
  } as any;
  const wrapper = mountWrapper(
    getStore({ repos: { repos: [repo] } }),
    <SelectRepoForm {...defaultProps} />,
  );
  const select = wrapper.find("select");
  select.simulate("change", { target: { value: "default/bitnami" } });
  expect(checkChart).toHaveBeenCalledWith(
    initialState.config.kubeappsCluster,
    repo.metadata.namespace,
    repo.metadata.name,
    defaultProps.chartName,
  );
});
