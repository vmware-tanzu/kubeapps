// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  PackageRepositoriesPermissions,
  PackageRepositoryDetail,
  PackageRepositorySummary,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories_pb";
import { getType } from "typesafe-actions";
import actions from "../actions";
import reposReducer, { IPackageRepositoryState } from "./repos";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins_pb";
import { PluginNames } from "shared/types";
import { Context } from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";

describe("reposReducer", () => {
  let initialState: IPackageRepositoryState;

  beforeEach(() => {
    initialState = {
      errors: {},
      isFetching: false,
      repoDetail: {} as PackageRepositoryDetail,
      reposSummaries: [] as PackageRepositorySummary[],
      reposPermissions: [] as PackageRepositoriesPermissions[],
    } as IPackageRepositoryState;
  });

  describe("repos", () => {
    const actionTypes = {
      addOrUpdateRepo: getType(actions.repos.addOrUpdateRepo),
      addedRepo: getType(actions.repos.addedRepo),
      errorRepos: getType(actions.repos.errorRepos),
      receiveRepo: getType(actions.repos.receiveRepoDetail),
      receiveRepos: getType(actions.repos.receiveRepoSummaries),
      repoUpdated: getType(actions.repos.repoUpdated),
      requestRepo: getType(actions.repos.requestRepoDetail),
      requestRepos: getType(actions.repos.requestRepoSummaries),
      requestReposPermissions: getType(actions.repos.requestReposPermissions),
      receiveReposPermissions: getType(actions.repos.receiveReposPermissions),
    };

    describe("reducer actions", () => {
      it("sets isFetching when requesting repos", () => {
        expect(
          reposReducer(undefined, {
            type: actionTypes.requestRepos,
            payload: "",
          }),
        ).toEqual({
          ...initialState,
          isFetching: true,
        });
      });

      it("unsets isFetching and receive repos", () => {
        const repoSummary = { name: "foo" } as PackageRepositorySummary;
        const state = reposReducer(undefined, {
          type: actionTypes.requestRepos,
          payload: "",
        });
        expect(state).toEqual({
          ...initialState,
          isFetching: true,
        });
        expect(
          reposReducer(state, {
            type: actionTypes.receiveRepos,
            payload: [repoSummary],
          }),
        ).toEqual({ ...initialState, reposSummaries: [repoSummary] } as IPackageRepositoryState);
      });

      it("unsets isFetching and receive repo", () => {
        const repoDetail = new PackageRepositoryDetail({ name: "foo" });
        const state = reposReducer(undefined, {
          type: actionTypes.requestRepos,
          payload: "",
        });
        expect(state).toEqual({
          ...initialState,
          isFetching: true,
        });
        expect(
          reposReducer(state, {
            type: actionTypes.receiveRepo,
            payload: repoDetail,
          }),
        ).toEqual({ ...initialState, repoDetail: repoDetail } as IPackageRepositoryState);
      });
    });

    it("adds a repo", () => {
      const repoSummary = { name: "foo" } as PackageRepositorySummary;
      const state = reposReducer(undefined, {
        type: actionTypes.addOrUpdateRepo,
      });
      expect(state).toEqual({
        ...initialState,
        isFetching: true,
      });
      expect(
        reposReducer(initialState, {
          type: actionTypes.addedRepo,
          payload: repoSummary,
        }),
      ).toEqual({ ...initialState, reposSummaries: [repoSummary] } as IPackageRepositoryState);
    });

    it("adds a repo sorting the result", () => {
      const repoSummary1 = { name: "zzz" } as PackageRepositorySummary;
      const repoSummary2 = { name: "aaa" } as PackageRepositorySummary;
      const state = { ...initialState, reposSummaries: [repoSummary1] } as IPackageRepositoryState;

      expect(
        reposReducer(state, {
          type: actionTypes.addedRepo,
          payload: repoSummary2,
        }),
      ).toEqual({
        ...initialState,
        reposSummaries: [repoSummary2, repoSummary1],
      } as IPackageRepositoryState);
    });

    it("updates a repo", () => {
      const repoSummary = new PackageRepositorySummary({
        name: "foo",
        url: "foo",
        packageRepoRef: { context: { namespace: "ns" } },
      });
      expect(
        reposReducer(
          { ...initialState, reposSummaries: [repoSummary] },
          {
            type: actionTypes.repoUpdated,
            payload: new PackageRepositorySummary({ ...repoSummary, url: "bar" }),
          },
        ),
      ).toEqual({
        ...initialState,
        reposSummaries: [{ ...repoSummary, url: "bar" }],
      } as IPackageRepositoryState);
    });

    it("unsets isFetching and receive repos permissions", () => {
      const plugin = { name: PluginNames.PACKAGES_HELM, version: "0.0.1" } as Plugin;
      const reposPermissions = [
        new PackageRepositoriesPermissions({
          plugin: plugin,
          global: {
            create: true,
            delete: true,
            list: true,
            update: true,
          },
          namespace: {
            create: true,
            delete: true,
            list: true,
            update: true,
          },
        }),
      ] as PackageRepositoriesPermissions[];
      const state = reposReducer(undefined, {
        type: actionTypes.requestReposPermissions,
        payload: { cluster: "", namespace: "" } as Context,
      });
      expect(state).toEqual({
        ...initialState,
        reposPermissions: [],
        isFetching: true,
      });
      expect(
        reposReducer(state, {
          type: actionTypes.receiveReposPermissions,
          payload: reposPermissions,
        }),
      ).toEqual({
        ...initialState,
        reposPermissions: reposPermissions,
        isFetching: false,
      } as IPackageRepositoryState);
    });
  });
});
