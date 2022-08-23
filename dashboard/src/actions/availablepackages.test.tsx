// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AvailablePackageDetail,
  AvailablePackageReference,
  AvailablePackageSummary,
  Context,
  GetAvailablePackageDetailResponse,
  GetAvailablePackageSummariesResponse,
  GetAvailablePackageVersionsResponse,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import PackagesService from "shared/PackagesService";
import { FetchError, IReceivePackagesActionPayload, IStoreState } from "shared/types";
import { getType } from "typesafe-actions";
import actions from ".";

const mockStore = configureMockStore([thunk]);

let store: any;

const namespace = "package-namespace";
const cluster = "default";
const repos = "foo";
const defaultPaginationToken = "defaultPaginationToken";
const defaultSize = 0;
const plugin = { name: "my.plugin", version: "0.0.1" } as Plugin;

const defaultAvailablePackageSummary: AvailablePackageSummary = {
  name: "foo",
  categories: [""],
  displayName: "foo",
  iconUrl: "",
  latestVersion: { appVersion: "v1.0.0", pkgVersion: "" },
  shortDescription: "",
  availablePackageRef: {
    identifier: "foo/foo",
    context: { cluster: "", namespace: "package-namespace" } as Context,
    plugin: plugin,
  },
};

const defaultAvailablePackageDetail: AvailablePackageDetail = {
  name: "foo",
  categories: [""],
  displayName: "foo",
  iconUrl: "",
  repoUrl: "",
  homeUrl: "",
  sourceUrls: [],
  shortDescription: "",
  longDescription: "",
  availablePackageRef: {
    identifier: "foo/foo",
    context: { cluster: "", namespace: "package-namespace" } as Context,
    plugin: plugin,
  },
  valuesSchema: "",
  defaultValues: "",
  maintainers: [],
  readme: "",
  version: {
    pkgVersion: "1.2.3",
    appVersion: "4.5.6",
  },
};

beforeEach(() => {
  store = mockStore({
    packages: {
      isFetching: false,
    },
  } as Partial<IStoreState>);
});

afterEach(() => {
  jest.restoreAllMocks();
});

interface IfetchAvailablePackageSummariesTestCase {
  name: string;
  response: GetAvailablePackageSummariesResponse;
  requestedRepos: string;
  requestedPageToken: string;
  requestedQuery?: string;
  expectedActions: any;
  expectedParams: any[];
}

const currentPageToken = "currentPageToken";
const nextPageToken = "nextPageToken";

const fetchAvailablePackageSummariesTestCases: IfetchAvailablePackageSummariesTestCase[] = [
  {
    name: "fetches packages with query",
    response: {
      availablePackageSummaries: [defaultAvailablePackageSummary],
      nextPageToken,
      categories: ["foo"],
    },
    requestedRepos: "",
    requestedPageToken: currentPageToken,
    requestedQuery: "foo",
    expectedActions: [
      {
        type: getType(actions.availablepackages.requestAvailablePackageSummaries),
        payload: currentPageToken,
      },
      {
        type: getType(actions.availablepackages.receiveAvailablePackageSummaries),
        payload: {
          response: {
            availablePackageSummaries: [defaultAvailablePackageSummary],
            nextPageToken,
            categories: ["foo"],
          },
          paginationToken: currentPageToken,
        } as IReceivePackagesActionPayload,
      },
    ],
    expectedParams: [cluster, namespace, "", currentPageToken, defaultSize, "foo"],
  },
  {
    name: "fetches packages from a repo (first page)",
    response: {
      availablePackageSummaries: [defaultAvailablePackageSummary],
      nextPageToken,
      categories: ["foo"],
    },
    requestedRepos: repos,
    requestedPageToken: "",
    expectedActions: [
      { type: getType(actions.availablepackages.requestAvailablePackageSummaries), payload: "" },
      {
        type: getType(actions.availablepackages.receiveAvailablePackageSummaries),
        payload: {
          response: {
            availablePackageSummaries: [defaultAvailablePackageSummary],
            nextPageToken,
            categories: ["foo"],
          },
          paginationToken: "",
        } as IReceivePackagesActionPayload,
      },
    ],
    expectedParams: [cluster, namespace, repos, "", defaultSize, undefined],
  },
  {
    name: "fetches packages from a repo (middle page)",
    response: {
      availablePackageSummaries: [defaultAvailablePackageSummary],
      nextPageToken,
      categories: ["foo"],
    },
    requestedRepos: repos,
    requestedPageToken: currentPageToken,
    expectedActions: [
      {
        type: getType(actions.availablepackages.requestAvailablePackageSummaries),
        payload: currentPageToken,
      },
      {
        type: getType(actions.availablepackages.receiveAvailablePackageSummaries),
        payload: {
          response: {
            availablePackageSummaries: [defaultAvailablePackageSummary],
            nextPageToken,
            categories: ["foo"],
          },
          paginationToken: currentPageToken,
        } as IReceivePackagesActionPayload,
      },
    ],
    expectedParams: [cluster, namespace, repos, currentPageToken, defaultSize, undefined],
  },
  {
    name: "fetches packages from a repo (last page)",
    response: {
      availablePackageSummaries: [defaultAvailablePackageSummary],
      nextPageToken: "",
      categories: ["foo"],
    },
    requestedRepos: repos,
    requestedPageToken: currentPageToken,
    expectedActions: [
      {
        type: getType(actions.availablepackages.requestAvailablePackageSummaries),
        payload: currentPageToken,
      },
      {
        type: getType(actions.availablepackages.receiveAvailablePackageSummaries),
        payload: {
          response: {
            availablePackageSummaries: [defaultAvailablePackageSummary],
            nextPageToken: "",
            categories: ["foo"],
          },
          paginationToken: currentPageToken,
        } as IReceivePackagesActionPayload,
      },
    ],
    expectedParams: [cluster, namespace, repos, currentPageToken, defaultSize, undefined],
  },
  {
    name: "fetches packages from a repo (already processed page)",
    response: {
      availablePackageSummaries: [defaultAvailablePackageSummary],
      nextPageToken,
      categories: ["foo"],
    },
    requestedRepos: repos,
    requestedPageToken: currentPageToken,
    expectedActions: [
      {
        type: getType(actions.availablepackages.requestAvailablePackageSummaries),
        payload: currentPageToken,
      },
      {
        type: getType(actions.availablepackages.receiveAvailablePackageSummaries),
        payload: {
          paginationToken: currentPageToken,
          response: {
            availablePackageSummaries: [defaultAvailablePackageSummary],
            nextPageToken,
            categories: ["foo"],
          },
        } as IReceivePackagesActionPayload,
      },
    ],
    expectedParams: [cluster, namespace, repos, currentPageToken, defaultSize, undefined],
  },
  {
    name: "fetches packages from a repo (off-limits page)",
    response: {
      availablePackageSummaries: [defaultAvailablePackageSummary],
      nextPageToken: "3",
      categories: ["foo"],
    },
    requestedRepos: repos,
    requestedPageToken: "next-page-token",
    expectedActions: [
      {
        type: getType(actions.availablepackages.requestAvailablePackageSummaries),
        payload: "next-page-token",
      },
      {
        type: getType(actions.availablepackages.receiveAvailablePackageSummaries),
        payload: {
          response: {
            availablePackageSummaries: [defaultAvailablePackageSummary],
            nextPageToken: "3",
            categories: ["foo"],
          },
          paginationToken: "next-page-token",
        } as IReceivePackagesActionPayload,
      },
    ],
    expectedParams: [cluster, namespace, repos, "next-page-token", defaultSize, undefined],
  },
];

describe("fetchAvailablePackageSummaries", () => {
  fetchAvailablePackageSummariesTestCases.forEach(tc => {
    it(tc.name, async () => {
      const mockGetAvailablePackageSummaries = jest
        .fn()
        .mockImplementation(() => Promise.resolve(tc.response));
      jest
        .spyOn(PackagesService, "getAvailablePackageSummaries")
        .mockImplementation(mockGetAvailablePackageSummaries);

      await store.dispatch(
        actions.availablepackages.fetchAvailablePackageSummaries(
          cluster,
          namespace,
          tc.requestedRepos,
          tc.requestedPageToken,
          defaultSize,
          tc.requestedQuery,
        ),
      );
      expect(store.getActions()).toEqual(tc.expectedActions);
      expect(mockGetAvailablePackageSummaries).toHaveBeenCalledWith(...tc.expectedParams);
    });
  });

  it("returns a 404 error", async () => {
    const expectedActions = [
      {
        type: getType(actions.availablepackages.requestAvailablePackageSummaries),
        payload: defaultPaginationToken,
      },
      {
        type: getType(actions.availablepackages.createErrorPackage),
        payload: new FetchError("could not find package"),
      },
    ];
    const mockGetAvailablePackageSummaries = jest.fn().mockImplementation(() => {
      throw new Error("could not find package");
    });
    jest
      .spyOn(PackagesService, "getAvailablePackageSummaries")
      .mockImplementation(mockGetAvailablePackageSummaries);
    await store.dispatch(
      actions.availablepackages.fetchAvailablePackageSummaries(
        cluster,
        namespace,
        "foo",
        defaultPaginationToken,
        defaultSize,
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns a generic error", async () => {
    const expectedActions = [
      {
        type: getType(actions.availablepackages.requestAvailablePackageSummaries),
        payload: defaultPaginationToken,
      },
      {
        type: getType(actions.availablepackages.createErrorPackage),
        payload: new Error("something went wrong"),
      },
    ];
    const mockGetAvailablePackageSummaries = jest.fn().mockImplementation(() => {
      throw new Error("something went wrong");
    });
    jest
      .spyOn(PackagesService, "getAvailablePackageSummaries")
      .mockImplementation(mockGetAvailablePackageSummaries);
    await store.dispatch(
      actions.availablepackages.fetchAvailablePackageSummaries(
        cluster,
        namespace,
        "foo",
        defaultPaginationToken,
        defaultSize,
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns a generic error and it is cleared later", async () => {
    const expectedActions = [
      {
        type: getType(actions.availablepackages.requestAvailablePackageSummaries),
        payload: defaultPaginationToken,
      },
      {
        type: getType(actions.availablepackages.createErrorPackage),
        payload: new Error("something went wrong"),
      },
      { type: getType(actions.availablepackages.clearErrorPackage) },
    ];
    const mockGetAvailablePackageSummaries = jest.fn().mockImplementation(() => {
      throw new Error("something went wrong");
    });
    jest
      .spyOn(PackagesService, "getAvailablePackageSummaries")
      .mockImplementation(mockGetAvailablePackageSummaries);
    await store.dispatch(
      actions.availablepackages.fetchAvailablePackageSummaries(
        cluster,
        namespace,
        "foo",
        defaultPaginationToken,
        defaultSize,
      ),
    );
    await store.dispatch(actions.availablepackages.clearErrorPackage());
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("fetchAvailablePackageVersions", () => {
  const packageAppVersions = [{ pkgVersion: "1.2.3", appVersion: "4.5.6" }];
  const availableVersionsResponse: GetAvailablePackageVersionsResponse = {
    packageAppVersions,
  };
  let mockGetAvailablePackageVersions: jest.Mock;
  beforeEach(() => {
    mockGetAvailablePackageVersions = jest
      .fn()
      .mockImplementation(() => Promise.resolve(availableVersionsResponse));
    jest
      .spyOn(PackagesService, "getAvailablePackageVersions")
      .mockImplementation(mockGetAvailablePackageVersions);
  });

  it("fetches package versions", async () => {
    const expectedActions = [
      { type: getType(actions.availablepackages.requestSelectedAvailablePackageVersions) },
      {
        type: getType(actions.availablepackages.receiveSelectedAvailablePackageVersions),
        payload: availableVersionsResponse,
      },
    ];
    await store.dispatch(
      actions.availablepackages.fetchAvailablePackageVersions({
        context: { cluster: cluster, namespace: namespace },
        identifier: "foo",
        plugin: plugin,
      } as AvailablePackageReference),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(mockGetAvailablePackageVersions.mock.calls[0]).toEqual([
      {
        context: { cluster: cluster, namespace: namespace },
        identifier: "foo",
        plugin: plugin,
      } as AvailablePackageReference,
    ]);
  });
});

describe("fetchAndSelectAvailablePackageDetail", () => {
  let mockGetAvailablePackageDetail: jest.Mock;
  beforeEach(() => {
    const response: GetAvailablePackageDetailResponse = {
      availablePackageDetail: defaultAvailablePackageDetail,
    };
    mockGetAvailablePackageDetail = jest.fn().mockImplementation(() => Promise.resolve(response));
    jest
      .spyOn(PackagesService, "getAvailablePackageDetail")
      .mockImplementation(mockGetAvailablePackageDetail);
  });

  it("gets a package version", async () => {
    const expectedActions = [
      { type: getType(actions.availablepackages.requestSelectedAvailablePackageDetail) },
      {
        type: getType(actions.availablepackages.receiveSelectedAvailablePackageDetail),
        payload: {
          selectedPackage: defaultAvailablePackageDetail,
        },
      },
    ];
    await store.dispatch(
      actions.availablepackages.fetchAndSelectAvailablePackageDetail(
        {
          context: { cluster: cluster, namespace: namespace },
          identifier: "foo",
          plugin: plugin,
        } as AvailablePackageReference,
        "1.0.0",
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(mockGetAvailablePackageDetail.mock.calls[0]).toEqual([
      {
        context: { cluster: cluster, namespace: namespace },
        identifier: "foo",
        plugin: plugin,
      } as AvailablePackageReference,
      "1.0.0",
    ]);
  });

  it("gets a package version with tag", async () => {
    const expectedActions = [
      { type: getType(actions.availablepackages.requestSelectedAvailablePackageDetail) },
      {
        type: getType(actions.availablepackages.receiveSelectedAvailablePackageDetail),
        payload: {
          selectedPackage: defaultAvailablePackageDetail,
        },
      },
    ];
    await store.dispatch(
      actions.availablepackages.fetchAndSelectAvailablePackageDetail(
        {
          context: { cluster: cluster, namespace: namespace },
          identifier: "foo",
          plugin: plugin,
        } as AvailablePackageReference,
        "1.0.0-alpha+1.2.3-beta2",
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(mockGetAvailablePackageDetail.mock.calls[0]).toEqual([
      {
        context: { cluster: cluster, namespace: namespace },
        identifier: "foo",
        plugin: plugin,
      } as AvailablePackageReference,

      "1.0.0-alpha+1.2.3-beta2",
    ]);
  });

  it("dispatches an error if it's unexpected", async () => {
    jest.spyOn(PackagesService, "getAvailablePackageDetail").mockImplementation(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      { type: getType(actions.availablepackages.requestSelectedAvailablePackageDetail) },
      { type: getType(actions.availablepackages.createErrorPackage), payload: new Error("Boom!") },
    ];
    await store.dispatch(
      actions.availablepackages.fetchAndSelectAvailablePackageDetail(
        {
          context: { cluster: cluster, namespace: namespace },
          identifier: "foo",
          plugin: plugin,
        } as AvailablePackageReference,
        "1.0.0",
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });
});
