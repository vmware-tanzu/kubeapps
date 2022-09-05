// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CanIRequest, CanIResponse } from "gen/kubeappsapis/plugins/resources/v1alpha1/resources";
import * as moxios from "moxios";
import { axiosWithAuth } from "./AxiosInstance";
import { Kube } from "./Kube";
import KubeappsGrpcClient from "./KubeappsGrpcClient";

const clusterName = "cluster-name";

describe("App", () => {
  beforeEach(() => {
    // Import as "any" to avoid typescript syntax error
    moxios.install(axiosWithAuth as any);
  });
  afterEach(() => {
    moxios.uninstall(axiosWithAuth as any);
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
    // Create a real client, but we'll stub out the function we're interested in.
    const client = new KubeappsGrpcClient().getResourcesServiceClientImpl();
    let mockClientCanI: jest.MockedFunction<typeof client.CanI>;

    beforeEach(() => {});
    afterEach(() => {
      jest.resetAllMocks();
    });

    it("should check permissions", async () => {
      mockClientCanI = jest
        .fn()
        .mockImplementation(() => Promise.resolve({ allowed: true } as CanIResponse));
      jest.spyOn(client, "CanI").mockImplementation(mockClientCanI);
      jest.spyOn(Kube, "resourcesServiceClient").mockImplementation(() => client);

      const allowed = await Kube.canI("cluster", "v1", "namespaces", "create", "");
      expect(allowed).toBe(true);

      expect(Kube.resourcesServiceClient).toHaveBeenCalledWith();
      expect(mockClientCanI).toHaveBeenCalledWith({
        context: {
          cluster: "cluster",
          namespace: "",
        },
        group: "v1",
        resource: "namespaces",
        verb: "create",
      } as CanIRequest);
    });
    it("should ignore empty clusters", async () => {
      const allowed = await Kube.canI("", "v1", "namespaces", "create", "");
      expect(allowed).toBe(false);
      expect(Kube.resourcesServiceClient).not.toHaveBeenCalled();
    });
    it("should default to disallow when errors", async () => {
      mockClientCanI = jest.fn().mockImplementation(
        () =>
          new Promise(() => {
            throw new Error("error");
          }),
      );
      jest.spyOn(client, "CanI").mockImplementation(mockClientCanI);
      jest.spyOn(Kube, "resourcesServiceClient").mockImplementation(() => client);

      const allowed = await Kube.canI("cluster", "v1", "secrets", "list", "");
      expect(allowed).toBe(false);

      expect(Kube.resourcesServiceClient).toHaveBeenCalled();
      expect(mockClientCanI).toHaveBeenCalledWith({
        context: {
          cluster: "cluster",
          namespace: "",
        },
        group: "v1",
        resource: "secrets",
        verb: "list",
      } as CanIRequest);
    });
  });
});
