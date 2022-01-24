import actions from "actions";
import Alert from "components/js/Alert";
import Table from "components/js/Table";
import Tooltip from "components/js/Tooltip";
import PageHeader from "components/PageHeader/PageHeader";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { Link } from "react-router-dom";
import { Kube } from "shared/Kube";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { AppRepoAddButton } from "./AppRepoButton";
import { AppRepoControl } from "./AppRepoControl";
import { AppRepoDisabledControl } from "./AppRepoDisabledControl";
import AppRepoList from "./AppRepoList";
import { AppRepoRefreshAllButton } from "./AppRepoRefreshAllButton";

const {
  clusters: { currentCluster, clusters },
  config: { globalReposNamespace },
} = initialState;
const namespace = clusters[currentCluster].currentNamespace;

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.repos = {
    ...actions.repos,
    fetchRepos: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  Kube.canI = jest.fn().mockReturnValue({
    then: jest.fn(f => f(true)),
  });
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("fetches repos and imagePullSecrets", () => {
  mountWrapper(defaultStore, <AppRepoList />);
  expect(actions.repos.fetchRepos).toHaveBeenCalledWith(namespace, true);
});

it("fetches repos only from the globalReposNamespace", () => {
  mountWrapper(
    getStore({
      clusters: {
        ...initialState.clusters,
        clusters: {
          [currentCluster]: {
            ...initialState.clusters.clusters[currentCluster],
            currentNamespace: globalReposNamespace,
          },
        },
      },
    }),
    <AppRepoList />,
  );
  expect(actions.repos.fetchRepos).toHaveBeenCalledWith(globalReposNamespace);
});

it("fetches repos from all namespaces (without kubeappsNamespace)", () => {
  const wrapper = mountWrapper(defaultStore, <AppRepoList />);
  act(() => {
    wrapper.find("input[type='checkbox']").simulate("change");
  });
  expect(actions.repos.fetchRepos).toHaveBeenCalledWith("");
});

it("should hide the all-namespace switch if the user doesn't have permissions", async () => {
  Kube.canI = jest.fn().mockReturnValue({
    then: jest.fn((f: any) => f(false)),
  });
  const wrapper = mountWrapper(defaultStore, <AppRepoList />);
  expect(wrapper.find("input[type='checkbox']")).not.toExist();
});

// TODO: Remove this test when app repos are supported in different clusters
it("shows a warning if the cluster is not the default one", () => {
  const wrapper = mountWrapper(
    getStore({
      clusters: {
        ...initialState.clusters,
        currentCluster: "other",
        clusters: {
          ...initialState.clusters.clusters,
          other: {
            ...initialState.clusters.clusters[currentCluster],
          },
        },
      },
    }),
    <AppRepoList />,
  );
  expect(wrapper.find(Alert)).toIncludeText(
    "App Repositories are available on the default cluster only",
  );
});

it("renders the button to add a repo and refresh all", () => {
  const wrapper = mountWrapper(defaultStore, <AppRepoList />);
  expect(wrapper.find(PageHeader).find(AppRepoAddButton)).toExist();
  expect(wrapper.find(PageHeader).find(AppRepoRefreshAllButton)).toExist();
});

it("shows an error fetching a repo", () => {
  const wrapper = mountWrapper(
    getStore({ repos: { errors: { fetch: new Error("boom!") } } }),
    <AppRepoList />,
  );
  expect(wrapper.find(Alert)).toIncludeText("boom!");
});

it("shows an error deleting a repo", () => {
  const wrapper = mountWrapper(
    getStore({ repos: { errors: { delete: new Error("boom!") } } }),
    <AppRepoList />,
  );
  expect(wrapper.find(Alert)).toIncludeText("boom!");
});

// TODO(andresmgot): Re-enable when the repo list is refactored
describe("global and namespaced repositories", () => {
  const globalRepo = {
    metadata: {
      name: "bitnami",
      namespace: globalReposNamespace,
    },
    spec: {},
  };

  const namespacedRepo = {
    metadata: {
      name: "my-repo",
      namespace,
    },
    spec: {
      description: "my description 1 2 3 4",
    },
  };

  it("shows a message if no global or namespaced repos exist", () => {
    const wrapper = mountWrapper(defaultStore, <AppRepoList />);
    expect(
      wrapper.find("p").filterWhere(p => p.text().includes("No global repositories found")),
    ).toExist();
    expect(
      wrapper
        .find("p")
        .filterWhere(p => p.text().includes("The current namespace doesn't have any repositories")),
    ).toExist();
  });

  it("shows the global repositories with the buttons deactivated if the current user is not allowed to modify them", () => {
    Kube.canI = jest.fn().mockReturnValue({
      then: jest.fn((f: any) => f(false)),
    });
    const wrapper = mountWrapper(
      getStore({
        clusters: {
          ...initialState.clusters,
          clusters: {
            [currentCluster]: {
              ...initialState.clusters.clusters[currentCluster],
              currentNamespace: "other",
            },
          },
        },
        repos: {
          repos: [globalRepo],
        },
      }),
      <AppRepoList />,
    );
    expect(wrapper.find(Table)).toHaveLength(1);
    // The control buttons should be deactivated
    expect(wrapper.find(AppRepoDisabledControl)).toExist();
    expect(wrapper.find(AppRepoControl)).not.toExist();
    // The content related to namespaced repositories should exist
    expect(
      wrapper.find("h3").filterWhere(h => h.text().includes("Namespace Repositories")),
    ).toExist();
  });

  it("shows the global repositories with the buttons enabled", () => {
    const wrapper = mountWrapper(
      getStore({
        clusters: {
          ...initialState.clusters,
          clusters: {
            [currentCluster]: {
              ...initialState.clusters.clusters[currentCluster],
              currentNamespace: globalReposNamespace,
            },
          },
        },
        repos: {
          repos: [globalRepo],
        },
      }),
      <AppRepoList />,
    );

    // A link to manage the repos should not exist since we are already there
    expect(wrapper.find("p").find(Link)).not.toExist();
    expect(wrapper.find(Table)).toHaveLength(1);
    // The control buttons should be enabled
    expect(wrapper.find(AppRepoDisabledControl)).not.toExist();
    expect(wrapper.find(AppRepoControl)).toExist();
    // The content related to namespaced repositories should be hidden
    expect(
      wrapper.find("h3").filterWhere(h => h.text().includes("Namespace Repositories")),
    ).not.toExist();
    // no tooltip for the global repo as it does not have a description.
    expect(wrapper.find(Tooltip)).not.toExist();
  });

  it("shows global and namespaced repositories", () => {
    const wrapper = mountWrapper(
      getStore({
        clusters: {
          ...initialState.clusters,
          clusters: {
            [currentCluster]: {
              ...initialState.clusters.clusters[currentCluster],
              currentNamespace: namespacedRepo.metadata.namespace,
            },
          },
        },
        repos: {
          repos: [globalRepo, namespacedRepo],
        },
      }),
      <AppRepoList />,
    );

    // A table per repository type
    expect(wrapper.find(Table)).toHaveLength(2);
  });

  it("shows a link to the repo catalog", () => {
    const wrapper = mountWrapper(
      getStore({
        repos: {
          repos: [namespacedRepo],
        },
      }),
      <AppRepoList />,
    );
    expect(wrapper.find(Table).find(Link).prop("to")).toEqual(
      `/c/${currentCluster}/ns/${namespacedRepo.metadata.namespace}/catalog?Repository=my-repo`,
    );
  });

  it("shows a tooltip for the repo", () => {
    const wrapper = mountWrapper(
      getStore({
        repos: {
          repos: [namespacedRepo],
        },
      }),
      <AppRepoList />,
    );
    act(() => {
      wrapper.find("input[type='checkbox']").simulate("change");
    });
    const tooltipText = wrapper.find(Tooltip).html();
    expect(tooltipText).toContain("my description 1 2 3 4");
  });

  it("use the correct namespace in the link when listing in all namespaces", () => {
    const wrapper = mountWrapper(
      getStore({
        repos: {
          repos: [namespacedRepo],
        },
      }),
      <AppRepoList />,
    );
    act(() => {
      wrapper.find("input[type='checkbox']").simulate("change");
    });
    expect(wrapper.find(Table).find(Link).prop("to")).toEqual(
      `/c/${currentCluster}/ns/${namespacedRepo.metadata.namespace}/catalog?Repository=my-repo`,
    );
  });
});
