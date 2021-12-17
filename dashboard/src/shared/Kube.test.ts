import * as moxios from "moxios";
import { axiosWithAuth } from "./AxiosInstance";
import { Kube } from "./Kube";

const clusterName = "cluster-name";

describe("App", () => {
  beforeEach(() => {
    // Import as "any" to avoid typescript syntax error
    moxios.install(axiosWithAuth as any);
  });
  afterEach(() => {
    moxios.uninstall(axiosWithAuth as any);
  });
  describe("getResourceURL", () => {
    [
      {
        description: "returns the version and resource",
        args: {
          cluster: clusterName,
          apiVersion: "v1",
          resource: "pods",
          namespaced: true,
        },
        result: `api/clusters/${clusterName}/api/v1/pods`,
      },
      {
        description: "defaults version to api/v1 if undefined",
        args: {
          cluster: clusterName,
          apiVersion: "",
          resource: "pods",
          namespaced: true,
        },
        result: `api/clusters/${clusterName}/api/v1/pods`,
      },
      {
        description: "skips the namespace if non-namespaced",
        args: {
          cluster: clusterName,
          apiVersion: "v1",
          resource: "clusterroles",
          namespaced: false,
          namespace: "default",
        },
        result: `api/clusters/${clusterName}/api/v1/clusterroles`,
      },
      {
        description: "returns the version, resource in a namespace",
        args: {
          cluster: clusterName,
          apiVersion: "",
          resource: "pods",
          namespaced: true,
          namespace: "default",
        },
        result: `api/clusters/${clusterName}/api/v1/namespaces/default/pods`,
      },
      {
        description: "returns the version, resource in a namespace with a name",
        args: {
          cluster: clusterName,
          apiVersion: "",
          resource: "pods",
          namespaced: true,
          namespace: "default",
          name: "foo",
        },
        result: `api/clusters/${clusterName}/api/v1/namespaces/default/pods/foo`,
      },
      {
        description: "returns the version, resource in a namespace with a name and a query",
        args: {
          cluster: clusterName,
          apiVersion: "",
          resource: "pods",
          namespaced: true,
          namespace: "default",
          name: "foo",
          label: "label=bar",
        },
        result: `api/clusters/${clusterName}/api/v1/namespaces/default/pods/foo?label=bar`,
      },
    ].forEach(t => {
      it(t.description, () => {
        expect(
          Kube.getResourceURL(
            t.args.cluster,
            t.args.apiVersion,
            t.args.resource,
            t.args.namespaced,
            t.args.namespace,
            t.args.name,
            t.args.label,
          ),
        ).toBe(t.result);
      });
    });
  });

  describe("getResource", () => {
    const resource = { name: "foo" };
    beforeEach(() => {
      moxios.stubRequest(/.*/, {
        response: { data: resource },
        status: 200,
      });
    });
    it("should request a resource", async () => {
      expect(
        await Kube.getResource(clusterName, "v1", "pods", true, "default", "foo", "label=bar"),
      ).toEqual({
        data: resource,
      });
      expect(moxios.requests.mostRecent().url).toBe(
        `api/clusters/${clusterName}/api/v1/namespaces/default/pods/foo?label=bar`,
      );
    });
  });

  describe("getAPIGroups", () => {
    const groups = [
      {
        name: "apiregistration.k8s.io",
        versions: [
          {
            groupVersion: "apiregistration.k8s.io/v1",
            version: "v1",
          },
        ],
        preferredVersion: {
          groupVersion: "apiregistration.k8s.io/v1",
          version: "v1",
        },
      },
    ];
    beforeEach(() => {
      moxios.stubRequest(/.*/, {
        // Sample response to /apis
        response: { kind: "APIGroupList", apiVersion: "v1", groups },
        status: 200,
      });
    });

    it("should request API groups", async () => {
      expect(await Kube.getAPIGroups(clusterName)).toEqual(groups);
      expect(moxios.requests.mostRecent().url).toBe(`api/clusters/${clusterName}/apis`);
    });
  });

  describe("getResourceKinds", () => {
    [
      {
        description: "returns v1 kinds",
        apiV1Response: {
          // Sample response to /api/v1
          response: {
            resources: [{ kind: "Pod", name: "pods", namespaced: true }],
          },
          status: 200,
        },
        groups: [],
        result: {
          Pod: {
            apiVersion: "v1",
            plural: "pods",
            namespaced: true,
          },
        },
      },
      {
        description: "returns additional groups kinds",
        apiV1Response: {
          // Sample response to /api/v1
          response: {
            resources: [{ kind: "Pod", name: "pods", namespaced: true }],
          },
          status: 200,
        },
        groups: [
          {
            input: { preferredVersion: { groupVersion: "rbac.authorization.k8s.io/v1" } },
            apiResponse: {
              response: {
                resources: [{ kind: "Role", name: "roles", namespaced: true }],
              },
            },
          },
        ],
        result: {
          Pod: {
            apiVersion: "v1",
            plural: "pods",
            namespaced: true,
          },
          Role: {
            apiVersion: "rbac.authorization.k8s.io/v1",
            plural: "roles",
            namespaced: true,
          },
        },
      },
      {
        description: "ignores subresources",
        apiV1Response: {
          // Sample response to /api/v1
          response: {
            resources: [
              { kind: "Pod", name: "pods", namespaced: true },
              { kind: "Pod", name: "pods/portforward", namespaced: true },
            ],
          },
          status: 200,
        },
        groups: [
          {
            input: { preferredVersion: { groupVersion: "rbac.authorization.k8s.io/v1" } },
            apiResponse: {
              response: {
                resources: [
                  { kind: "Role", name: "roles", namespaced: true },
                  { kind: "Role", name: "roles/other", namespaced: true },
                ],
              },
            },
          },
        ],
        result: {
          Pod: {
            apiVersion: "v1",
            plural: "pods",
            namespaced: true,
          },
          Role: {
            apiVersion: "rbac.authorization.k8s.io/v1",
            plural: "roles",
            namespaced: true,
          },
        },
      },
    ].forEach(t => {
      it(t.description, async () => {
        // eslint-disable-next-line redos/no-vulnerable
        moxios.stubRequest(/.*api\/v1/, t.apiV1Response);
        const groups: any[] = [];
        t.groups.forEach((g: any) => {
          groups.push(g.input);
          // eslint-disable-next-line redos/no-vulnerable
          moxios.stubOnce("GET", /.*apis\/.*/, g.apiResponse);
        });
        expect(await Kube.getResourceKinds("cluster", groups)).toEqual(t.result);
      });
    });
  });

  describe("canI", () => {
    beforeEach(() => {
      moxios.stubRequest(/.*/, {
        response: { allowed: true },
        status: 200,
      });
    });
    it("should check permissions", async () => {
      const allowed = await Kube.canI("cluster", "v1", "namespaces", "create", "");
      expect(allowed).toBe(true);
    });
    it("should ignore empty clusters", async () => {
      const allowed = await Kube.canI("", "v1", "namespaces", "create", "");
      expect(allowed).toBe(false);
    });
  });
});
