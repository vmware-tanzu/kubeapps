import Alert from "components/js/Alert";
import * as React from "react";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { IAppRepository } from "shared/types";
import SelectRepoForm from "./SelectRepoForm.v2";

const defaultProps = {
  isFetching: false,
  cluster: "default",
  namespace: "kubeapps",
  repo: {} as IAppRepository,
  repos: [] as IAppRepository[],
  chartName: "test",
  checkChart: jest.fn(),
  fetchRepositories: jest.fn(),
};

it("should fetch repositories", () => {
  const fetch = jest.fn();
  mountWrapper(defaultStore, <SelectRepoForm {...defaultProps} fetchRepositories={fetch} />);
  expect(fetch).toHaveBeenCalled();
});

it("should render a loading page if fetching", () => {
  expect(
    mountWrapper(defaultStore, <SelectRepoForm {...defaultProps} />).find("LoadingWrapper"),
  ).toExist();
});

it("render an error if failed to request repos", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <SelectRepoForm {...defaultProps} repoError={new Error("boom")} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("boom");
});

it("render a warning if there are no repos", () => {
  const wrapper = mountWrapper(defaultStore, <SelectRepoForm {...defaultProps} repos={[]} />);
  expect(wrapper.find(Alert)).toIncludeText("Chart repositories not found");
});

it("should select a repo", () => {
  const checkChart = jest.fn();
  const repo = {
    metadata: {
      name: "bitnami",
    },
    spec: {
      url: "http://repo",
    },
  } as any;
  const wrapper = mountWrapper(
    defaultStore,
    <SelectRepoForm {...defaultProps} repos={[repo]} checkChart={checkChart} />,
  );
  const select = wrapper.find("select");
  select.simulate("change", { target: { value: "bitnami" } });
  expect(checkChart).toHaveBeenCalledWith(
    defaultProps.namespace,
    "bitnami",
    defaultProps.chartName,
  );
});
