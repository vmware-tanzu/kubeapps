import * as moxios from "moxios";
import { axiosWithAuth } from "./AxiosInstance";
import Chart from "./Chart";

const clusterName = "cluster-name";
const namespaceName = "namespace-name";
const defaultPage = 1;
const defaultSize = 0;

describe("App", () => {
  beforeEach(() => {
    // Import as "any" to avoid typescript syntax error
    moxios.install(axiosWithAuth as any);
    moxios.stubRequest(/.*/, {
      response: { data: "ok" },
      status: 200,
    });
  });
  afterEach(() => {
    moxios.uninstall(axiosWithAuth as any);
  });
  describe("fetchCharts", () => {
    [
      {
        description: "fetch charts url without repos, without query",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
          repos: "",
          page: defaultPage,
          size: defaultSize,
          query: "",
        },
        result: `api/assetsvc/v1/clusters/${clusterName}/namespaces/${namespaceName}/charts?page=${defaultPage}&size=${defaultSize}`,
      },
      {
        description: "fetch charts url without repos, with query",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
          repos: "",
          page: defaultPage,
          size: defaultSize,
          query: "cms",
        },
        result: `api/assetsvc/v1/clusters/${clusterName}/namespaces/${namespaceName}/charts?page=${defaultPage}&size=${defaultSize}&q=cms`,
      },
      {
        description: "fetch charts url with repos, without query",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
          repos: "repo1,repo2",
          page: defaultPage,
          size: defaultSize,
          query: "",
        },
        result: `api/assetsvc/v1/clusters/${clusterName}/namespaces/${namespaceName}/charts?page=${defaultPage}&size=${defaultSize}&repos=repo1,repo2`,
      },
      {
        description: "fetch charts url wtih repos, with query",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
          repos: "repo1,repo2",
          page: defaultPage,
          size: defaultSize,
          query: "cms",
        },
        result: `api/assetsvc/v1/clusters/${clusterName}/namespaces/${namespaceName}/charts?page=${defaultPage}&size=${defaultSize}&q=cms&repos=repo1,repo2`,
      },
    ].forEach(t => {
      it(t.description, async () => {
        expect(
          await Chart.fetchCharts(
            t.args.cluster,
            t.args.namespace,
            t.args.repos,
            t.args.page,
            t.args.size,
            t.args.query,
          ),
        ).toStrictEqual({ data: "ok" });
        expect(moxios.requests.mostRecent().url).toStrictEqual(t.result);
      });
    });
  });
  describe("fetchChartCategories", () => {
    [
      {
        description: "fetch chart categories url",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
        },
        result: `api/assetsvc/v1/clusters/${clusterName}/namespaces/${namespaceName}/charts/categories`,
      },
    ].forEach(t => {
      it(t.description, async () => {
        expect(await Chart.fetchChartCategories(t.args.cluster, t.args.namespace)).toStrictEqual(
          "ok",
        );
        expect(moxios.requests.mostRecent().url).toStrictEqual(t.result);
      });
    });
  });
  describe("fetchChartVersions", () => {
    [
      {
        description: "fetch chart versions url",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
          id: "mychart",
        },
        result: `api/assetsvc/v1/clusters/${clusterName}/namespaces/${namespaceName}/charts/mychart/versions`,
      },
    ].forEach(t => {
      it(t.description, async () => {
        expect(
          await Chart.fetchChartVersions(t.args.cluster, t.args.namespace, t.args.id),
        ).toStrictEqual("ok");
        expect(moxios.requests.mostRecent().url).toStrictEqual(t.result);
      });
    });
  });
  describe("fetchChartVersion", () => {
    [
      {
        description: "fetch chart version url",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
          id: "mychart",
          version: "v1",
        },
        result: `api/assetsvc/v1/clusters/${clusterName}/namespaces/${namespaceName}/charts/mychart/versions/v1`,
      },
    ].forEach(t => {
      it(t.description, async () => {
        expect(
          await Chart.getChartVersion(t.args.cluster, t.args.namespace, t.args.id, t.args.version),
        ).toStrictEqual("ok");
        expect(moxios.requests.mostRecent().url).toStrictEqual(t.result);
      });
    });
  });
  describe("getReadme", () => {
    [
      {
        description: "get readme url",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
          id: "mychart",
          version: "v1",
        },
        result: `api/assetsvc/v1/clusters/${clusterName}/namespaces/${namespaceName}/assets/mychart/versions/v1/README.md`,
      },
    ].forEach(t => {
      it(t.description, async () => {
        expect(
          await Chart.getReadme(t.args.cluster, t.args.namespace, t.args.id, t.args.version),
        ).toStrictEqual({ data: "ok" });
        expect(moxios.requests.mostRecent().url).toStrictEqual(t.result);
      });
    });
  });
  describe("getValues", () => {
    [
      {
        description: "get values url",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
          id: "mychart",
          version: "v1",
        },
        result: `api/assetsvc/v1/clusters/${clusterName}/namespaces/${namespaceName}/assets/mychart/versions/v1/values.yaml`,
      },
    ].forEach(t => {
      it(t.description, async () => {
        expect(
          await Chart.getValues(t.args.cluster, t.args.namespace, t.args.id, t.args.version),
        ).toStrictEqual({ data: "ok" });
        expect(moxios.requests.mostRecent().url).toStrictEqual(t.result);
      });
    });
  });
  describe("getSchema", () => {
    [
      {
        description: "get schema url",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
          id: "mychart",
          version: "v1",
        },
        result: `api/assetsvc/v1/clusters/${clusterName}/namespaces/${namespaceName}/assets/mychart/versions/v1/values.schema.json`,
      },
    ].forEach(t => {
      it(t.description, async () => {
        expect(
          await Chart.getSchema(t.args.cluster, t.args.namespace, t.args.id, t.args.version),
        ).toStrictEqual({ data: "ok" });
        expect(moxios.requests.mostRecent().url).toStrictEqual(t.result);
      });
    });
  });
  describe("listWithFilters", () => {
    [
      {
        description: "listWithFilters url",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
          name: "mychart",
          version: "v1",
          appVersion: "1.0.1",
        },
        result: `api/assetsvc/v1/clusters/${clusterName}/namespaces/${namespaceName}/charts?name=mychart&version=v1&appversion=1.0.1`,
      },
    ].forEach(t => {
      it(t.description, async () => {
        expect(
          await Chart.listWithFilters(
            t.args.cluster,
            t.args.namespace,
            t.args.name,
            t.args.version,
            t.args.appVersion,
          ),
        ).toStrictEqual("ok");
        expect(moxios.requests.mostRecent().url).toStrictEqual(t.result);
      });
    });
  });
});
