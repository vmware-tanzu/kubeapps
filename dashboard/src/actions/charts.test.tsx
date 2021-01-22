import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";
import { axiosWithAuth } from "../shared/AxiosInstance";

import actions from ".";
import {
  FetchError,
  IChart,
  IChartListMeta,
  IReceiveChartsActionPayload,
  NotFoundError,
} from "../shared/types";

const mockStore = configureMockStore([thunk]);

let axiosGetMock = jest.fn();
let store: any;
let response: any;

const namespace = "chart-namespace";
const cluster = "default";
const defaultPage = 1;
const defaultSize = 0;
let defaultRecords = new Map<number, boolean>().set(1, false);

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
  defaultRecords = new Map<number, boolean>().set(1, false);
});

afterEach(() => {
  jest.resetAllMocks();
});

describe("fetchCharts", () => {
  it("fetches charts from a repo (first page)", async () => {
    response = { data: [chartItem] as IChart[], meta: { totalPages: 1 } as IChartListMeta };
    const expectedActions = [
      { type: getType(actions.charts.requestCharts), payload: 1 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          items: response.data,
          page: 1,
          totalPages: response.meta.totalPages,
        } as IReceiveChartsActionPayload,
      },
    ];
    axiosWithAuth.get = axiosGetMock;
    await store.dispatch(
      actions.charts.fetchCharts(
        cluster,
        namespace,
        "foo",
        defaultPage,
        defaultSize,
        defaultRecords,
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts?page=${defaultPage}&size=${defaultSize}&repos=foo`,
    );
  });

  it("fetches charts from a repo (middle page)", async () => {
    response = { data: [chartItem] as IChart[], meta: { totalPages: 3 } as IChartListMeta };
    const expectedActions = [
      { type: getType(actions.charts.requestCharts), payload: 2 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          items: response.data,
          page: 2,
          totalPages: response.meta.totalPages,
        } as IReceiveChartsActionPayload,
      },
    ];
    axiosWithAuth.get = axiosGetMock;
    await store.dispatch(
      actions.charts.fetchCharts(
        cluster,
        namespace,
        "foo",
        2,
        defaultSize,
        defaultRecords.set(1, true).set(2, false),
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts?page=${2}&size=${defaultSize}&repos=foo`,
    );
  });

  it("fetches charts from a repo (last page)", async () => {
    response = { data: [chartItem] as IChart[], meta: { totalPages: 3 } as IChartListMeta };
    const expectedActions = [
      { type: getType(actions.charts.requestCharts), payload: 3 },
      {
        type: getType(actions.charts.receiveCharts),
        payload: {
          items: response.data,
          page: 3,
          totalPages: response.meta.totalPages,
        } as IReceiveChartsActionPayload,
      },
    ];
    axiosWithAuth.get = axiosGetMock;
    await store.dispatch(
      actions.charts.fetchCharts(
        cluster,
        namespace,
        "foo",
        3,
        defaultSize,
        defaultRecords
          .set(1, true)
          .set(2, true)
          .set(3, false),
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}/charts?page=${3}&size=${defaultSize}&repos=foo`,
    );
  });

  it("fetches charts from a repo (already processed page)", async () => {
    response = { data: [chartItem] as IChart[], meta: { totalPages: 3 } as IChartListMeta };
    const expectedActions = [] as any;
    axiosWithAuth.get = axiosGetMock;
    await store.dispatch(
      actions.charts.fetchCharts(
        cluster,
        namespace,
        "foo",
        2, // request page 2
        defaultSize,
        defaultRecords
          .set(1, true)
          .set(2, true)
          .set(3, false), // but the next one should be page 3
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls).toHaveLength(0);
  });

  it("fetches charts from a repo (not-yet-requested page)", async () => {
    response = { data: [chartItem] as IChart[], meta: { totalPages: 3 } as IChartListMeta };
    const expectedActions = [] as any;
    axiosWithAuth.get = axiosGetMock;
    await store.dispatch(
      actions.charts.fetchCharts(
        cluster,
        namespace,
        "foo",
        4, // request page 4
        defaultSize,
        defaultRecords
          .set(1, true)
          .set(2, true)
          .set(3, false), // but the next one should be page 3
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls).toHaveLength(0);
  });

  it("fetches charts from a repo (already-requested page)", async () => {
    response = { data: [chartItem] as IChart[], meta: { totalPages: 3 } as IChartListMeta };
    const expectedActions = [] as any;
    axiosWithAuth.get = axiosGetMock;
    await store.dispatch(
      actions.charts.fetchCharts(
        cluster,
        namespace,
        "foo",
        2, // request page 2
        defaultSize,
        defaultRecords
          .set(1, true)
          .set(2, true)
          .set(3, false), // but the next one should be page 3
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls).toHaveLength(0);
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
      actions.charts.fetchCharts(
        cluster,
        namespace,
        "foo",
        defaultPage,
        defaultSize,
        defaultRecords,
      ),
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
      actions.charts.fetchCharts(
        cluster,
        namespace,
        "foo",
        defaultPage,
        defaultSize,
        defaultRecords,
      ),
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
