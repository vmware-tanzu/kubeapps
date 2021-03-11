import { CdsButton } from "@cds/react/button";
import actions from "actions";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import { AppRepoRefreshAllButton } from "./AppRepoRefreshAllButton";

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.repos = {
    ...actions.repos,
    resyncAllRepos: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("refreshes all repos", () => {
  const repos = [
    { metadata: { name: "foo", namespace: "default" } },
    { metadata: { name: "bar", namespace: "kubeapps" } },
  ];
  const wrapper = mountWrapper(getStore({ repos: { repos } }), <AppRepoRefreshAllButton />);
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();
  expect(actions.repos.resyncAllRepos).toHaveBeenCalledWith([
    { name: "foo", namespace: "default" },
    { name: "bar", namespace: "kubeapps" },
  ]);
});
