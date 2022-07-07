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
      addingRepo: false,
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
            type: actionTypes.requestRepos,
            payload: "",
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
          type: actionTypes.requestRepos,
          payload: "",
        });
        expect(state).toEqual({
          ...initialState,
          isFetching: true,
          isFetchingElem: { repositories: true, secrets: false },
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
          isFetchingElem: { repositories: true, secrets: false },
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
        type: actionTypes.addRepo,
      });
      expect(state).toEqual({
        ...initialState,
        addingRepo: true,
      });
      expect(
        reposReducer(state, {
          type: actionTypes.addedRepo,
          payload: repoSummary,
        }),
      ).toEqual({ ...initialState, repos: [repoSummary] });
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
