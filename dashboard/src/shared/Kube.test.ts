// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import MockAdapter from "axios-mock-adapter";
import {
  CanIRequest,
  CanIResponse,
} from "gen/kubeappsapis/plugins/resources/v1alpha1/resources_pb";
import { axiosWithAuth } from "./AxiosInstance";
import { Kube } from "./Kube";
import KubeappsGrpcClient from "./KubeappsGrpcClient";

const clusterName = "cluster-name";

describe("Kube", () => {
  let axiosMock: MockAdapter;

  beforeEach(() => {
    axiosMock = new MockAdapter(axiosWithAuth);
  });
  afterEach(() => {
    axiosMock.restore();
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

    it("should request API groups", async () => {
      axiosMock.onGet().reply(200, { kind: "APIGroupList", apiVersion: "v1", groups });
      expect(await Kube.getAPIGroups(clusterName)).toEqual(groups);
      const request = axiosMock.history.get[0];
      expect(request?.url).toBe(`api/clusters/${clusterName}/apis`);
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
        axiosMock.onGet(/.*api\/v1/).reply(t.apiV1Response.status, t.apiV1Response.response);
        const groups: any[] = [];
        t.groups.forEach((group: any) => {
          groups.push(group.input);
          // eslint-disable-next-line redos/no-vulnerable
          axiosMock.onGet(/.*apis\/.*/).replyOnce(200, group.apiResponse.response);
        });
        expect(await Kube.getResourceKinds(clusterName, groups)).toEqual(t.result);
      });
    });
  });

  describe("canI", () => {
    // Create a real client, but we'll stub out the function we're interested in.
    const client = new KubeappsGrpcClient().getResourcesServiceClientImpl();
    let mockClientCanI: jest.MockedFunction<typeof client.canI>;

    beforeEach(() => {});
    afterEach(() => {
      jest.resetAllMocks();
    });

    it("should check permissions", async () => {
      mockClientCanI = jest
        .fn()
        .mockImplementation(() => Promise.resolve({ allowed: true } as CanIResponse));
      jest.spyOn(client, "canI").mockImplementation(mockClientCanI);
      jest.spyOn(Kube, "resourcesServiceClient").mockImplementation(() => client);

      const allowed = await Kube.canI(clusterName, "v1", "namespaces", "create", "");
      expect(allowed).toBe(true);

      expect(Kube.resourcesServiceClient).toHaveBeenCalledWith();
      expect(mockClientCanI).toHaveBeenCalledWith({
        context: {
          cluster: clusterName,
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
      jest.spyOn(client, "canI").mockImplementation(mockClientCanI);
      jest.spyOn(Kube, "resourcesServiceClient").mockImplementation(() => client);

      const allowed = await Kube.canI(clusterName, "v1", "secrets", "list", "");
      expect(allowed).toBe(false);

      expect(Kube.resourcesServiceClient).toHaveBeenCalled();
      expect(mockClientCanI).toHaveBeenCalledWith({
        context: {
          cluster: clusterName,
          namespace: "",
        },
        group: "v1",
        resource: "secrets",
        verb: "list",
      } as CanIRequest);
    });
  });
});
