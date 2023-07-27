// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import ResourceRef, { fromCRD } from "./ResourceRef";
import { IClusterServiceVersionCRDResource } from "./types";

const clusterName = "cluster-name";

describe("ResourceRef", () => {
  describe("constructor", () => {
    it("returns a ResourceRef with the correct details", () => {
      const ref = new ResourceRef(clusterName, "deployments", true, "releaseNamespace", {
        apiVersion: "apps/v1",
        kind: "Deployment",
        name: "foo",
        namespace: "bar",
      });
      expect(ref).toBeInstanceOf(ResourceRef);
      expect(ref).toEqual({
        cluster: clusterName,
        apiVersion: "apps/v1",
        kind: "Deployment",
        name: "foo",
        namespace: "bar",
        namespaced: true,
        plural: "deployments",
      });
    });

    it("sets a default namespace if not in the resource", () => {
      const ref = new ResourceRef(clusterName, "deployments", true, "default", {
        apiVersion: "apps/v1",
        kind: "Deployment",
        name: "foo",
      });
      expect(ref.namespace).toBe("default");
    });

    describe("fromCRD", () => {
      it("creates a resource ref with ownerReference", () => {
        const r = {
          kind: "Deployment",
          name: "",
          version: "",
        } as IClusterServiceVersionCRDResource;
        const ownerRef = {
          metadata: {
            name: "test",
          },
        };
        const kind = {
          apiVersion: "apps/v1",
          plural: "deployments",
          namespaced: true,
        };
        const res = fromCRD(r, kind, clusterName, "default", ownerRef);
        expect(res).toMatchObject({
          apiVersion: "apps/v1",
          kind: "Deployment",
          name: "",
          namespace: "default",
          filter: { metadata: { ownerReferences: [ownerRef] } },
        });
      });

      it("skips the namespace for a non namespaced element", () => {
        const r = {
          kind: "ClusterRole",
          name: "",
          version: "",
        } as IClusterServiceVersionCRDResource;
        const ownerRef = {
          metadata: {
            name: "test",
          },
        };
        const kind = {
          apiVersion: "rbac.authorization.k8s.io/v1",
          plural: "clusterroles",
          namespaced: false,
        };
        const res = fromCRD(r, kind, clusterName, "default", ownerRef);
        expect(res).toMatchObject({
          apiVersion: "rbac.authorization.k8s.io/v1",
          kind: "ClusterRole",
          name: "",
          namespace: "",
          filter: { metadata: { ownerReferences: [ownerRef] } },
        });
      });
    });
  });
});
