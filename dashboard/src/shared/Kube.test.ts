// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

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
