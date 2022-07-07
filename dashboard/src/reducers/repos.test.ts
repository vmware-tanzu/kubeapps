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
      form: {
        name: "",
        namespace: "",
        show: false,
        url: "",
      },
      isFetching: false,
      isFetchingElem: {
        repositories: false,
        secrets: false,
      },
      validating: false,
      repo: {} as PackageRepositoryDetail,
      repos: [] as PackageRepositorySummary[],
      imagePullSecrets: [],
    };
  });

  describe("repos", () => {
    const actionTypes = {
      addRepo: getType(actions.repos.addRepo),
      addedRepo: getType(actions.repos.addedRepo),
      requestRepoUpdate: getType(actions.repos.requestRepoUpdate),
      repoUpdated: getType(actions.repos.repoUpdated),
      requestRepos: getType(actions.repos.requestRepoSummaries),
      receiveRepos: getType(actions.repos.receiveRepoSummaries),
      requestRepo: getType(actions.repos.requestRepoDetail),
      receiveRepo: getType(actions.repos.receiveRepoDetail),
      redirect: getType(actions.repos.redirect),
      redirected: getType(actions.repos.redirected),
      errorRepos: getType(actions.repos.errorRepos),
    };

    describe("reducer actions", () => {
      it("sets isFetching when requesting repos", () => {
        expect(
          reposReducer(undefined, {
            type: actionTypes.requestRepos as any,
          }),
        ).toEqual({
          ...initialState,
          isFetching: true,
          isFetchingElem: { repositories: true, secrets: false },
        });
      });

      it("unsets isFetching and receive repos", () => {
        const repoSummary = { name: "foo" } as PackageRepositorySummary;
        const state = reposReducer(undefined, {
          type: actionTypes.requestRepos as any,
        });
        expect(state).toEqual({
          ...initialState,
          isFetching: true,
          isFetchingElem: { repositories: true, secrets: false },
        });
        expect(
          reposReducer(state, {
            type: actionTypes.receiveRepos as any,
            payload: [repoSummary],
          }),
        ).toEqual({ ...initialState, repos: [repoSummary] });
      });

      it("unsets isFetching and receive repo", () => {
        const repoDetail = { name: "foo" } as PackageRepositoryDetail;
        const state = reposReducer(undefined, {
          type: actionTypes.requestRepos as any,
        });
        expect(state).toEqual({
          ...initialState,
          isFetching: true,
          isFetchingElem: { repositories: true, secrets: false },
        });
        expect(
          reposReducer(state, {
            type: actionTypes.receiveRepo as any,
            payload: repoDetail,
          }),
        ).toEqual({ ...initialState, repo: repoDetail });
      });
    });

    it("adds a repo", () => {
      const repoDetail = { name: "foo" } as PackageRepositoryDetail;
      const state = reposReducer(undefined, {
        type: actionTypes.addRepo as any,
      });
      expect(state).toEqual({
        ...initialState,
      });
      expect(
        reposReducer(state, {
          type: actionTypes.addedRepo as any,
          payload: repoDetail,
        }),
      ).toEqual({ ...initialState, lastAdded: repoDetail, repo: repoDetail });
    });

    it("updates a repo", () => {
      const repoDetail = { name: "foo" } as PackageRepositoryDetail;
      expect(
        reposReducer(
          { ...initialState, repo: repoDetail },
          {
            type: actionTypes.repoUpdated as any,
            payload: { ...repoDetail, name: "b" },
          },
        ),
      ).toEqual({ ...initialState, repo: { ...repoDetail, name: "b" } });
    });
  });
});
