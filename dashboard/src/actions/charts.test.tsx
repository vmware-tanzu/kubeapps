import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";
import { axiosWithAuth } from "../shared/AxiosInstance";

import actions from ".";
import { FetchError, IChart, IChartListMeta, NotFoundError } from "../shared/types";

const mockStore = configureMockStore([thunk]);

let axiosGetMock = jest.fn();
let store: any;
let response: any;

const namespace = "chart-namespace";
const cluster = "default";
const repos = "foo";
const defaultPage = 1;
const defaultSize = 0;

const chartItem = {
  id: "foo",
  attributes: {
    name: "foo",
    description: "",
    category: "",
    repo: { name: "foo", namespace: "chart-namespace" },
  },
  relationships: { latestChartVersion: { data: { app_version: "v1.0.0" } } },
} as IChart;

beforeEach(() => {
  store = mockStore();
  axiosGetMock.mockImplementation(() => {
    return {
      status: 200,
      data: response,
    };
  });
  axiosWithAuth.get = axiosGetMock;
});

afterEach(() => {
  jest.resetAllMocks();
});

interface IFetchChartsTestCase {
  name: string;
  responseData: IChart[];
  responseMeta: IChartListMeta;
  requestedRepos: string;
  requestedPage: number;
  requestedQuery?: string;
  expectedActions: any;
  expectedURL: string;
}

const fetchChartsTestCases: IFetchChartsTestCase[] = [
  {
    name: "fetches charts with query",
    responseData: [chartItem],
    responseMeta: { totalPages: 1 },
    requestedRepos: "",
    requestedPage: 1,
    requestedQuery: "foo",
    expectedActions: [
      { type: getType(actions.charts.requestCharts), payload: 1 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          items: [chartItem],
          page: 1,
          totalPages: 1,
        },
      },
    ],
    expectedURL: `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts?page=${defaultPage}&size=${defaultSize}&q=foo`,
  },
  {
    name: "fetches charts from a repo (first page)",
    responseData: [chartItem],
    responseMeta: { totalPages: 3 },
    requestedRepos: repos,
    requestedPage: 1,
    expectedActions: [
      { type: getType(actions.charts.requestCharts), payload: 1 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          items: [chartItem],
          page: 1,
          totalPages: 3,
        },
      },
    ],
    expectedURL: `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts?page=${1}&size=${defaultSize}&repos=foo`,
  },
  {
    name: "fetches charts from a repo (middle page)",
    responseData: [chartItem],
    responseMeta: { totalPages: 3 },
    requestedRepos: repos,
    requestedPage: 2,
    expectedActions: [
      { type: getType(actions.charts.requestCharts), payload: 2 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          items: [chartItem],
          page: 2,
          totalPages: 3,
        },
      },
    ],
    expectedURL: `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts?page=${2}&size=${defaultSize}&repos=${repos}`,
  },
  {
    name: "fetches charts from a repo (last page)",
    responseData: [chartItem],
    responseMeta: { totalPages: 3 },
    requestedRepos: repos,
    requestedPage: 3,
    expectedActions: [
      { type: getType(actions.charts.requestCharts), payload: 3 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          items: [chartItem],
          page: 3,
          totalPages: 3,
        },
      },
    ],
    expectedURL: `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts?page=${3}&size=${defaultSize}&repos=${repos}`,
  },
  {
    name: "fetches charts from a repo (already processed page)",
    responseData: [chartItem],
    responseMeta: { totalPages: 3 },
    requestedRepos: repos,
    requestedPage: 2,
    expectedActions: [
      { type: getType(actions.charts.requestCharts), payload: 2 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          items: [chartItem],
          page: 2,
          totalPages: 3,
        },
      },
    ],
    expectedURL: `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts?page=${2}&size=${defaultSize}&repos=${repos}`,
  },
  {
    name: "fetches charts from a repo (off-limits page)",
    responseData: [chartItem],
    responseMeta: { totalPages: 3 },
    requestedRepos: repos,
    requestedPage: 4,
    expectedActions: [
      { type: getType(actions.charts.requestCharts), payload: 4 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          items: [chartItem],
          page: 4,
          totalPages: 3,
        },
      },
    ],
    expectedURL: `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts?page=${4}&size=${defaultSize}&repos=${repos}`,
  },
];

describe("fetchCharts", () => {
  fetchChartsTestCases.forEach(tc => {
    it(tc.name, async () => {
      response = { data: tc.responseData, meta: tc.responseMeta };
      axiosWithAuth.get = axiosGetMock;
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
      expect(axiosGetMock.mock.calls[0][0]).toBe(tc.expectedURL);
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
    axiosGetMock = jest.fn(() => {
      throw new Error("could not find chart");
    });
    axiosWithAuth.get = axiosGetMock;
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
    axiosGetMock = jest.fn(() => {
      throw new Error("something went wrong");
    });
    axiosWithAuth.get = axiosGetMock;
    await store.dispatch(
      actions.charts.fetchCharts(cluster, namespace, "foo", defaultPage, defaultSize),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("fetchChartCategories", () => {
  it("fetches chart categories", async () => {
    response = { data: [{ id: "foo" }] };
    const expectedActions = [
      { type: getType(actions.charts.requestChartsCategories) },
      { type: getType(actions.charts.receiveChartCategories), payload: response.data },
    ];
    await store.dispatch(actions.charts.fetchChartCategories(cluster, namespace));
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts/categories`,
    );
  });

  it("returns a 404 error", async () => {
    const expectedActions = [
      { type: getType(actions.charts.requestChartsCategories) },
      {
        type: getType(actions.charts.errorChartCatetories),
        payload: new FetchError("could not find chart categories"),
      },
    ];
    axiosGetMock = jest.fn(() => {
      throw new Error("could not find chart categories");
    });
    axiosWithAuth.get = axiosGetMock;
    await store.dispatch(actions.charts.fetchChartCategories(cluster, namespace));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns a generic error", async () => {
    const expectedActions = [
      { type: getType(actions.charts.requestChartsCategories) },
      {
        type: getType(actions.charts.errorChartCatetories),
        payload: new Error("something went wrong"),
      },
    ];
    axiosGetMock = jest.fn(() => {
      throw new Error("something went wrong");
    });
    axiosWithAuth.get = axiosGetMock;
    await store.dispatch(actions.charts.fetchChartCategories(cluster, namespace));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("fetchChartVersions", () => {
  it("fetches chart versions", async () => {
    response = { data: [{ id: "foo" }] };
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.receiveChartVersions), payload: response.data },
    ];
    await store.dispatch(actions.charts.fetchChartVersions(cluster, namespace, "foo"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts/foo/versions`,
    );
  });
});

describe("getChartVersion", () => {
  it("gets a chart version", async () => {
    response = { data: { id: "foo" } };
    const expectedActions = [
      { type: getType(actions.charts.requestChart) },
      {
        type: getType(actions.charts.selectChartVersion),
        payload: {
          chartVersion: response.data,
          schema: { data: response.data },
          values: { data: response.data },
        },
      },
    ];
    await store.dispatch(actions.charts.getChartVersion(cluster, namespace, "foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts/foo/versions/1.0.0`,
    );
  });

  it("gets a chart version with tag", async () => {
    response = { data: { id: "foo" } };
    const expectedActions = [
      { type: getType(actions.charts.requestChart) },
      {
        type: getType(actions.charts.selectChartVersion),
        payload: {
          chartVersion: response.data,
          schema: { data: response.data },
          values: { data: response.data },
        },
      },
    ];
    await store.dispatch(
      actions.charts.getChartVersion(cluster, namespace, "foo", "1.0.0-alpha+1.2.3-beta2"),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts/foo/versions/1.0.0-alpha%2B1.2.3-beta2`,
    );
  });

  it("gets a chart version with values and schema", async () => {
    // Call to get the chart version
    axiosGetMock.mockImplementationOnce(() => {
      return {
        status: 200,
        data: { data: { id: "foo" } },
      };
    });
    // Call to get the chart values
    axiosGetMock.mockImplementationOnce(() => {
      return {
        status: 200,
        data: "foo: bar",
      };
    });
    // Call to get the chart schema
    axiosGetMock.mockImplementationOnce(() => {
      return {
        status: 200,
        data: { properties: "foo" },
      };
    });
    const expectedActions = [
      { type: getType(actions.charts.requestChart) },
      {
        type: getType(actions.charts.selectChartVersion),
        payload: { chartVersion: { id: "foo" }, values: "foo: bar", schema: { properties: "foo" } },
      },
    ];
    await store.dispatch(actions.charts.getChartVersion(cluster, namespace, "foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns an empty schema if not found", async () => {
    // Call to get the chart version
    axiosGetMock.mockImplementationOnce(() => {
      return {
        status: 200,
        data: {
          data: { id: "foo" },
        },
      };
    });
    // Call to get the chart values
    axiosGetMock.mockImplementationOnce(() => {
      throw new NotFoundError();
    });
    // Call to get the chart schema
    axiosGetMock.mockImplementationOnce(() => {
      throw new NotFoundError();
    });
    const expectedActions = [
      { type: getType(actions.charts.requestChart) },
      {
        type: getType(actions.charts.selectChartVersion),
        payload: { chartVersion: { id: "foo" }, values: "", schema: {} },
      },
    ];
    await store.dispatch(actions.charts.getChartVersion(cluster, namespace, "foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error if it's unexpected", async () => {
    // Call to get the chart version
    axiosGetMock.mockImplementationOnce(() => {
      return {
        status: 200,
        data: {
          data: { id: "foo" },
        },
      };
    });
    // Call to get the chart values
    axiosGetMock.mockImplementationOnce(() => {
      throw new Error("Boom!");
    });
    // Call to get the chart schema
    axiosGetMock.mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      { type: getType(actions.charts.requestChart) },
      { type: getType(actions.charts.errorChart), payload: new Error("Boom!") },
    ];
    await store.dispatch(actions.charts.getChartVersion(cluster, namespace, "foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("fetchChartVersionsAndSelectVersion", () => {
  it("fetches charts and select a version", async () => {
    response = { data: [{ id: "foo", attributes: { version: "1.0.0" } }] };
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.receiveChartVersions), payload: response.data },
      {
        type: getType(actions.charts.selectChartVersion),
        payload: { chartVersion: response.data[0] },
      },
    ];
    await store.dispatch(
      actions.charts.fetchChartVersionsAndSelectVersion(cluster, namespace, "foo", "1.0.0"),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts/foo/versions`,
    );
  });

  it("returns a not found error", async () => {
    response = { data: [{ id: "foo", attributes: { version: "1.0.0" } }] };
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      {
        type: getType(actions.charts.errorChart),
        payload: new FetchError("could not find chart"),
      },
    ];
    axiosGetMock = jest.fn(() => {
      throw new Error("could not find chart");
    });
    axiosWithAuth.get = axiosGetMock;
    await store.dispatch(
      actions.charts.fetchChartVersionsAndSelectVersion(cluster, namespace, "foo", "1.0.0"),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts/foo/versions`,
    );
  });

  it("selects the latest version by default", async () => {
    response = {
      data: [
        { id: "foo", attributes: { version: "1.0.0" } },
        { id: "foo", attributes: { version: "1.0.0" } },
      ],
    };
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.receiveChartVersions), payload: response.data },
      {
        type: getType(actions.charts.selectChartVersion),
        payload: { chartVersion: response.data[1] },
      },
    ];
    await store.dispatch(
      actions.charts.fetchChartVersionsAndSelectVersion(cluster, namespace, "foo", ""),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts/foo/versions`,
    );
  });
});

describe("getDeployedChartVersion", () => {
  it("should request a deployed chart", async () => {
    response = { data: { id: "foo" } };
    const expectedActions = [
      { type: getType(actions.charts.requestDeployedChartVersion) },
      {
        type: getType(actions.charts.receiveDeployedChartVersion),
        payload: {
          chartVersion: response.data,
          schema: { data: response.data },
          values: { data: response.data },
        },
      },
    ];
    await store.dispatch(
      actions.charts.getDeployedChartVersion(cluster, namespace, "foo", "1.0.0"),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });
});
