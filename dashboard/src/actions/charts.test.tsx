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
import Chart from "shared/Chart";
import { FetchError, IReceiveChartsActionPayload } from "shared/types";
import { getType } from "typesafe-actions";
import actions from ".";

const mockStore = configureMockStore([thunk]);

let store: any;

const namespace = "chart-namespace";
const cluster = "default";
const repos = "foo";
const defaultPage = 1;
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
    context: { cluster: "", namespace: "chart-namespace" } as Context,
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
    context: { cluster: "", namespace: "chart-namespace" } as Context,
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
  store = mockStore();
});

afterEach(() => {
  jest.restoreAllMocks();
});

interface IFetchChartsTestCase {
  name: string;
  response: GetAvailablePackageSummariesResponse;
  requestedRepos: string;
  requestedPage: number;
  requestedQuery?: string;
  expectedActions: any;
  expectedParams: any[];
}

const fetchChartsTestCases: IFetchChartsTestCase[] = [
  {
    name: "fetches charts with query",
    response: {
      availablePackageSummaries: [defaultAvailablePackageSummary],
      nextPageToken: "1",
      categories: ["foo"],
    },
    requestedRepos: "",
    requestedPage: 1,
    requestedQuery: "foo",
    expectedActions: [
      { type: getType(actions.charts.requestCharts), payload: 1 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          response: {
            availablePackageSummaries: [defaultAvailablePackageSummary],
            nextPageToken: "1",
            categories: ["foo"],
          },
          page: 1,
        } as IReceiveChartsActionPayload,
      },
    ],
    expectedParams: [cluster, namespace, "", 1, defaultSize, "foo"],
  },
  {
    name: "fetches charts from a repo (first page)",
    response: {
      availablePackageSummaries: [defaultAvailablePackageSummary],
      nextPageToken: "3",
      categories: ["foo"],
    },
    requestedRepos: repos,
    requestedPage: 1,
    expectedActions: [
      { type: getType(actions.charts.requestCharts), payload: 1 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          response: {
            availablePackageSummaries: [defaultAvailablePackageSummary],
            nextPageToken: "3",
            categories: ["foo"],
          },
          page: 1,
        } as IReceiveChartsActionPayload,
      },
    ],
    expectedParams: [cluster, namespace, repos, 1, defaultSize, undefined],
  },
  {
    name: "fetches charts from a repo (middle page)",
    response: {
      availablePackageSummaries: [defaultAvailablePackageSummary],
      nextPageToken: "3",
      categories: ["foo"],
    },
    requestedRepos: repos,
    requestedPage: 2,
    expectedActions: [
      { type: getType(actions.charts.requestCharts), payload: 2 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          response: {
            availablePackageSummaries: [defaultAvailablePackageSummary],
            nextPageToken: "3",
            categories: ["foo"],
          },
          page: 2,
        } as IReceiveChartsActionPayload,
      },
    ],
    expectedParams: [cluster, namespace, repos, 2, defaultSize, undefined],
  },
  {
    name: "fetches charts from a repo (last page)",
    response: {
      availablePackageSummaries: [defaultAvailablePackageSummary],
      nextPageToken: "3",
      categories: ["foo"],
    },
    requestedRepos: repos,
    requestedPage: 3,
    expectedActions: [
      { type: getType(actions.charts.requestCharts), payload: 3 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          response: {
            availablePackageSummaries: [defaultAvailablePackageSummary],
            nextPageToken: "3",
            categories: ["foo"],
          },
          page: 3,
        } as IReceiveChartsActionPayload,
      },
    ],
    expectedParams: [cluster, namespace, repos, 3, defaultSize, undefined],
  },
  {
    name: "fetches charts from a repo (already processed page)",
    response: {
      availablePackageSummaries: [defaultAvailablePackageSummary],
      nextPageToken: "3",
      categories: ["foo"],
    },
    requestedRepos: repos,
    requestedPage: 2,
    expectedActions: [
      { type: getType(actions.charts.requestCharts), payload: 2 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          response: {
            availablePackageSummaries: [defaultAvailablePackageSummary],
            nextPageToken: "3",
            categories: ["foo"],
          },
          page: 2,
        } as IReceiveChartsActionPayload,
      },
    ],
    expectedParams: [cluster, namespace, repos, 2, defaultSize, undefined],
  },
  {
    name: "fetches charts from a repo (off-limits page)",
    response: {
      availablePackageSummaries: [defaultAvailablePackageSummary],
      nextPageToken: "3",
      categories: ["foo"],
    },
    requestedRepos: repos,
    requestedPage: 4,
    expectedActions: [
      { type: getType(actions.charts.requestCharts), payload: 4 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          response: {
            availablePackageSummaries: [defaultAvailablePackageSummary],
            nextPageToken: "3",
            categories: ["foo"],
          },
          page: 4,
        } as IReceiveChartsActionPayload,
      },
    ],
    expectedParams: [cluster, namespace, repos, 4, defaultSize, undefined],
  },
];

describe("fetchCharts", () => {
  fetchChartsTestCases.forEach(tc => {
    it(tc.name, async () => {
      const mockGetAvailablePackageSummaries = jest
        .fn()
        .mockImplementation(() => Promise.resolve(tc.response));
      jest
        .spyOn(Chart, "getAvailablePackageSummaries")
        .mockImplementation(mockGetAvailablePackageSummaries);

      await store.dispatch(
        actions.charts.fetchCharts(
          cluster,
          namespace,
          tc.requestedRepos,
          tc.requestedPage,
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
      { type: getType(actions.charts.requestCharts), payload: 1 },
      {
        type: getType(actions.charts.errorChart),
        payload: new FetchError("could not find chart"),
      },
    ];
    const mockGetAvailablePackageSummaries = jest.fn().mockImplementation(() => {
      throw new Error("could not find chart");
    });
    jest
      .spyOn(Chart, "getAvailablePackageSummaries")
      .mockImplementation(mockGetAvailablePackageSummaries);
    await store.dispatch(
      actions.charts.fetchCharts(cluster, namespace, "foo", defaultPage, defaultSize),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns a generic error", async () => {
    const expectedActions = [
      { type: getType(actions.charts.requestCharts), payload: 1 },
      { type: getType(actions.charts.errorChart), payload: new Error("something went wrong") },
    ];
    const mockGetAvailablePackageSummaries = jest.fn().mockImplementation(() => {
      throw new Error("something went wrong");
    });
    jest
      .spyOn(Chart, "getAvailablePackageSummaries")
      .mockImplementation(mockGetAvailablePackageSummaries);
    await store.dispatch(
      actions.charts.fetchCharts(cluster, namespace, "foo", defaultPage, defaultSize),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns a generic error and it is cleared later", async () => {
    const expectedActions = [
      { type: getType(actions.charts.requestCharts), payload: 1 },
      { type: getType(actions.charts.errorChart), payload: new Error("something went wrong") },
      { type: getType(actions.charts.clearErrorChart) },
    ];
    const mockGetAvailablePackageSummaries = jest.fn().mockImplementation(() => {
      throw new Error("something went wrong");
    });
    jest
      .spyOn(Chart, "getAvailablePackageSummaries")
      .mockImplementation(mockGetAvailablePackageSummaries);
    await store.dispatch(
      actions.charts.fetchCharts(cluster, namespace, "foo", defaultPage, defaultSize),
    );
    await store.dispatch(actions.charts.clearErrorChart());
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("fetchChartVersions", () => {
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
      .spyOn(Chart, "getAvailablePackageVersions")
      .mockImplementation(mockGetAvailablePackageVersions);
  });

  it("fetches chart versions", async () => {
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.receiveChartVersions), payload: availableVersionsResponse },
    ];
    await store.dispatch(
      actions.charts.fetchChartVersions({
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

describe("fetchChartVersion", () => {
  let mockGetAvailablePackageDetail: jest.Mock;
  beforeEach(() => {
    const response: GetAvailablePackageDetailResponse = {
      availablePackageDetail: defaultAvailablePackageDetail,
    };
    mockGetAvailablePackageDetail = jest.fn().mockImplementation(() => Promise.resolve(response));
    jest
      .spyOn(Chart, "getAvailablePackageDetail")
      .mockImplementation(mockGetAvailablePackageDetail);
  });

  it("gets a chart version", async () => {
    const expectedActions = [
      {
        type: getType(actions.charts.selectChartVersion),
        payload: {
          selectedPackage: defaultAvailablePackageDetail,
        },
      },
    ];
    await store.dispatch(
      actions.charts.fetchChartVersion(
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

  it("gets a chart version with tag", async () => {
    const expectedActions = [
      {
        type: getType(actions.charts.selectChartVersion),
        payload: {
          selectedPackage: defaultAvailablePackageDetail,
        },
      },
    ];
    await store.dispatch(
      actions.charts.fetchChartVersion(
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
    jest.spyOn(Chart, "getAvailablePackageDetail").mockImplementation(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      { type: getType(actions.charts.errorChart), payload: new Error("Boom!") },
    ];
    await store.dispatch(
      actions.charts.fetchChartVersion(
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

describe("getDeployedChartVersion", () => {
  it("should request a deployed chart", async () => {
    const response: GetAvailablePackageDetailResponse = {
      availablePackageDetail: defaultAvailablePackageDetail,
    };
    const mockGetAvailablePackageDetail = jest
      .fn()
      .mockImplementation(() => Promise.resolve(response));
    jest
      .spyOn(Chart, "getAvailablePackageDetail")
      .mockImplementation(mockGetAvailablePackageDetail);

    const expectedActions = [
      { type: getType(actions.charts.requestDeployedChartVersion) },
      {
        type: getType(actions.charts.receiveDeployedChartVersion),
        payload: {
          chartVersion: defaultAvailablePackageDetail,
          schema: defaultAvailablePackageDetail.valuesSchema,
          values: defaultAvailablePackageDetail.defaultValues,
        },
      },
    ];

    await store.dispatch(
      actions.charts.getDeployedChartVersion(
        {
          context: { cluster: cluster, namespace: namespace },
          identifier: "foo",
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        } as AvailablePackageReference,
        "1.0.0",
      ),
    );

    expect(store.getActions()).toEqual(expectedActions);
    expect(mockGetAvailablePackageDetail.mock.calls[0]).toEqual([
      {
        context: { cluster: cluster, namespace: namespace },
        identifier: "foo",
        plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
      } as AvailablePackageReference,
      "1.0.0",
    ]);
  });
});
