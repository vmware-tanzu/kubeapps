import { getType } from "typesafe-actions";
import actions from "../actions";

import { IAppRepository } from "shared/types";
import reposReducer, { IAppRepositoryState } from "./repos";

describe("reposReducer", () => {
  let initialState: IAppRepositoryState;

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
      repo: {} as IAppRepository,
      repos: [],
      repoSecrets: [],
      imagePullSecrets: [],
    };
  });

  describe("repos", () => {
    const actionTypes = {
      addRepo: getType(actions.repos.addRepo),
      addedRepo: getType(actions.repos.addedRepo),
      requestRepoUpdate: getType(actions.repos.requestRepoUpdate),
      repoUpdated: getType(actions.repos.repoUpdated),
      requestRepos: getType(actions.repos.requestRepos),
      receiveRepos: getType(actions.repos.receiveRepos),
      receiveReposSecrets: getType(actions.repos.receiveReposSecrets),
      receiveReposSecret: getType(actions.repos.receiveReposSecret),
      requestRepo: getType(actions.repos.requestRepo),
      receiveRepo: getType(actions.repos.receiveRepo),
      repoValidating: getType(actions.repos.repoValidating),
      repoValidated: getType(actions.repos.repoValidated),
      clearRepo: getType(actions.repos.clearRepo),
      showForm: getType(actions.repos.showForm),
      hideForm: getType(actions.repos.hideForm),
      resetForm: getType(actions.repos.resetForm),
      submitForm: getType(actions.repos.submitForm),
      redirect: getType(actions.repos.redirect),
      redirected: getType(actions.repos.redirected),
      errorRepos: getType(actions.repos.errorRepos),
      requestImagePullSecrets: getType(actions.repos.requestImagePullSecrets),
      receiveImagePullSecrets: getType(actions.repos.receiveImagePullSecrets),
      createImagePullSecret: getType(actions.repos.createImagePullSecret),
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
        const repo = { metadata: { name: "foo" } };
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
            payload: [repo],
          }),
        ).toEqual({ ...initialState, repos: [repo] });
      });

      it("unsets isFetching and receive repo", () => {
        const repo = { metadata: { name: "foo" } };
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
            payload: repo,
          }),
        ).toEqual({ ...initialState, repo });
      });
    });

    it("receives repo secrets", () => {
      const secret = { metadata: { name: "foo" } };
      expect(
        reposReducer(undefined, {
          type: actionTypes.receiveReposSecrets as any,
          payload: [secret],
        }),
      ).toEqual({ ...initialState, repoSecrets: [secret] });
    });

    it("receives a repo secret", () => {
      const secret = { metadata: { name: "foo" } } as any;
      const modifiedSecret = { ...secret, spec: { foo: "bar" } };
      expect(
        reposReducer(
          { ...initialState, repoSecrets: [secret] },
          {
            type: actionTypes.receiveReposSecret as any,
            payload: modifiedSecret,
          },
        ),
      ).toEqual({ ...initialState, repoSecrets: [modifiedSecret] });
    });

    it("adds a repo", () => {
      const repo = { metadata: { name: "foo" } };
      const state = reposReducer(undefined, {
        type: actionTypes.addRepo as any,
      });
      expect(state).toEqual({
        ...initialState,
        addingRepo: true,
      });
      expect(
        reposReducer(state, {
          type: actionTypes.addedRepo as any,
          payload: repo,
        }),
      ).toEqual({ ...initialState, lastAdded: repo, repos: [repo] });
    });

    it("updates a repo", () => {
      const repo = { metadata: { name: "foo" } } as any;
      expect(
        reposReducer(
          { ...initialState, repos: [repo] },
          {
            type: actionTypes.repoUpdated as any,
            payload: { ...repo, spec: { a: "b" } },
          },
        ),
      ).toEqual({ ...initialState, repos: [{ ...repo, spec: { a: "b" } }] });
    });

    it("validates a repo", () => {
      const state = reposReducer(undefined, {
        type: actionTypes.repoValidating as any,
      });
      expect(state).toEqual({
        ...initialState,
        validating: true,
      });
      expect(
        reposReducer(state, {
          type: actionTypes.repoValidated as any,
        }),
      ).toEqual({ ...initialState });
    });

    it("receives image pull secrets", () => {
      const pullSecret = { metadata: { name: "foo" } } as any;
      const state = reposReducer(undefined, {
        type: actionTypes.requestImagePullSecrets as any,
      });
      expect(state).toEqual({
        ...initialState,
        isFetching: true,
        isFetchingElem: { repositories: false, secrets: true },
      });
      expect(
        reposReducer(state, {
          type: actionTypes.receiveImagePullSecrets as any,
          payload: [pullSecret],
        }),
      ).toEqual({ ...initialState, imagePullSecrets: [pullSecret] });
    });
  });
});
