// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  PackageRepositoryDetail,
  PackageRepositorySummary,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { getType } from "typesafe-actions";
import actions from "../actions";
import reposReducer, { IPackageRepositoryState } from "./repos";

describe("reposReducer", () => {
  let initialState: IPackageRepositoryState;

  beforeEach(() => {
    initialState = {
      errors: {},
      isFetching: false,
      repo: {} as PackageRepositoryDetail,
      repos: [] as PackageRepositorySummary[],
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
        ).toEqual({ ...initialState, repos: [repoSummary] });
      });

      it("unsets isFetching and receive repo", () => {
        const repoDetail = { name: "foo" } as PackageRepositoryDetail;
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
        ).toEqual({ ...initialState, repo: repoDetail });
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
      ).toEqual({ ...initialState, repos: [repoSummary] });
    });

    it("adds a repo sorting the result", () => {
      const repoSummary1 = { name: "zzz" } as PackageRepositorySummary;
      const repoSummary2 = { name: "aaa" } as PackageRepositorySummary;
      const state = { ...initialState, repos: [repoSummary1] };

      expect(
        reposReducer(state, {
          type: actionTypes.addedRepo,
          payload: repoSummary2,
        }),
      ).toEqual({ ...initialState, repos: [repoSummary2, repoSummary1] });
    });

    it("updates a repo", () => {
      const repoSummary = {
        name: "foo",
        url: "foo",
        packageRepoRef: { context: { namespace: "ns" } },
      } as PackageRepositorySummary;
      expect(
        reposReducer(
          { ...initialState, repos: [repoSummary] },
          {
            type: actionTypes.repoUpdated,
            payload: { ...repoSummary, url: "bar" },
          },
        ),
      ).toEqual({ ...initialState, repos: [{ ...repoSummary, url: "bar" }] });
    });
  });
});
