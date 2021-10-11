import { AvailablePackageSummary, Context } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { getType } from "typesafe-actions";
import actions from "../actions";
import { IPackageState, IReceivePackagesActionPayload } from "../shared/types";
import packageReducer from "./packages";

describe("packageReducer", () => {
  let initialState: IPackageState;
  const availablePackageSummary1: AvailablePackageSummary = {
    name: "foo",
    categories: [""],
    displayName: "foo",
    iconUrl: "",
    latestVersion: { appVersion: "v1.0.0", pkgVersion: "" },
    shortDescription: "",
    availablePackageRef: {
      identifier: "foo/foo",
      context: { cluster: "", namespace: "package-namespace" } as Context,
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    },
  };

  const availablePackageSummary2: AvailablePackageSummary = {
    name: "bar",
    categories: ["Database"],
    displayName: "bar",
    iconUrl: "",
    latestVersion: { appVersion: "v2.0.0", pkgVersion: "" },
    shortDescription: "",
    availablePackageRef: {
      identifier: "bar/bar",
      context: { cluster: "", namespace: "package-namespace" } as Context,
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    },
  };

  beforeEach(() => {
    initialState = {
      isFetching: false,
      hasFinishedFetching: false,
      items: [],
      categories: [],
      selected: {
        versions: [],
      },
      size: 20,
    };
  });
  const error = new Error("Boom");

  it("unsets an error when changing namespace", () => {
    const state = packageReducer(undefined, {
      type: getType(actions.packages.createErrorPackage) as any,
      payload: error,
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      selected: {
        ...initialState.selected,
        error,
      },
    });

    expect(
      packageReducer(undefined, {
        type: getType(actions.namespace.setNamespaceState) as any,
      }),
    ).toEqual({ ...initialState });
  });

  it("requestAvailablePackageSummaries (without page)", () => {
    const state = packageReducer(undefined, {
      type: getType(actions.packages.requestAvailablePackageSummaries) as any,
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: true,
    });
  });

  it("requestAvailablePackageSummaries (with page)", () => {
    const state = packageReducer(undefined, {
      type: getType(actions.packages.requestAvailablePackageSummaries) as any,
      payload: 2,
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: true,
    });
  });

  it("single receiveAvailablePackageSummaries (first page) should be returned", () => {
    const state = packageReducer(undefined, {
      type: getType(actions.packages.receiveAvailablePackageSummaries) as any,
      payload: {
        response: {
          availablePackageSummaries: [availablePackageSummary1],
          nextPageToken: "3",
          categories: ["foo"],
        },
        page: 1,
      } as IReceivePackagesActionPayload,
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: false,
      categories: ["foo"],
      items: [availablePackageSummary1],
    });
  });

  it("single receiveAvailablePackageSummaries (middle page) having visited the previous ones should be ignored", () => {
    const state = packageReducer(
      { ...initialState },
      {
        type: getType(actions.packages.receiveAvailablePackageSummaries) as any,
        payload: {
          response: {
            availablePackageSummaries: [availablePackageSummary1],
            nextPageToken: "3",
            categories: ["foo"],
          },
          page: 2,
        } as IReceivePackagesActionPayload,
      },
    );
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: false,
      categories: ["foo"],
      items: [availablePackageSummary1],
    });
  });

  it("single receiveAvailablePackageSummaries (middle page) not visiting the previous ones should be ignored", () => {
    const state = packageReducer(
      { ...initialState },
      {
        type: getType(actions.packages.receiveAvailablePackageSummaries) as any,
        payload: {
          response: {
            availablePackageSummaries: [availablePackageSummary1],
            nextPageToken: "3",
            categories: ["foo"],
          },
          page: 2,
        } as IReceivePackagesActionPayload,
      },
    );
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: false,
      categories: ["foo"],
      items: [availablePackageSummary1],
    });
  });

  it("single receiveAvailablePackageSummaries (last page) not incrementing page", () => {
    const state = packageReducer(
      {
        ...initialState,
      },
      {
        type: getType(actions.packages.receiveAvailablePackageSummaries) as any,
        payload: {
          response: {
            availablePackageSummaries: [availablePackageSummary1],
            nextPageToken: "3",
            categories: ["foo"],
          },
          page: 3,
        } as IReceivePackagesActionPayload,
      },
    );
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: true,
      categories: ["foo"],
      items: [availablePackageSummary1],
    });
  });

  it("two receiveAvailablePackageSummaries should add items (no dups)", () => {
    const state1 = packageReducer(undefined, {
      type: getType(actions.packages.receiveAvailablePackageSummaries) as any,
      payload: {
        response: {
          availablePackageSummaries: [availablePackageSummary1],
          nextPageToken: "2",
          categories: ["foo"],
        },
        page: 1,
      } as IReceivePackagesActionPayload,
    });
    const state2 = packageReducer(state1, {
      type: getType(actions.packages.receiveAvailablePackageSummaries) as any,
      payload: {
        response: {
          availablePackageSummaries: [availablePackageSummary2],
          nextPageToken: "2",
          categories: ["foo"],
        },
        page: 2,
      } as IReceivePackagesActionPayload,
    });
    expect(state2).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: true,
      categories: ["foo"],
      items: [availablePackageSummary1, availablePackageSummary2],
    });
    expect(state2.items.length).toBe(2);
  });

  it("requestAvailablePackageSummaries and receiveAvailablePackageSummaries with multiple pages", () => {
    const stateReq1 = packageReducer(initialState, {
      type: getType(actions.packages.requestAvailablePackageSummaries) as any,
      payload: 1,
    });
    expect(stateReq1).toEqual({
      ...initialState,
      isFetching: true,
      hasFinishedFetching: false,
      items: [],
    });
    const stateRec1 = packageReducer(stateReq1, {
      type: getType(actions.packages.receiveAvailablePackageSummaries) as any,
      payload: {
        response: {
          availablePackageSummaries: [availablePackageSummary1],
          nextPageToken: "3",
          categories: ["foo"],
        },
        page: 1,
      } as IReceivePackagesActionPayload,
    });
    expect(stateRec1).toEqual({
      ...initialState,
      isFetching: false,
      categories: ["foo"],
      items: [availablePackageSummary1],
      hasFinishedFetching: false,
    });
    const stateReq2 = packageReducer(stateRec1, {
      type: getType(actions.packages.requestAvailablePackageSummaries) as any,
      payload: 2,
    });
    expect(stateReq2).toEqual({
      ...initialState,
      isFetching: true,
      hasFinishedFetching: false,
      categories: ["foo"],
      items: [availablePackageSummary1],
    });
    const stateRec2 = packageReducer(stateReq2, {
      type: getType(actions.packages.receiveAvailablePackageSummaries) as any,
      payload: {
        response: {
          availablePackageSummaries: [availablePackageSummary2],
          nextPageToken: "3",
          categories: ["foo"],
        },
        page: 2,
      } as IReceivePackagesActionPayload,
    });
    expect(stateRec2).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: false,
      categories: ["foo"],
      items: [availablePackageSummary1, availablePackageSummary2],
    });
    const stateReq3 = packageReducer(stateRec2, {
      type: getType(actions.packages.requestAvailablePackageSummaries) as any,
      payload: 3,
    });
    expect(stateReq3).toEqual({
      ...initialState,
      isFetching: true,
      hasFinishedFetching: false,
      categories: ["foo"],
      items: [availablePackageSummary1, availablePackageSummary2],
    });
    const stateRec3 = packageReducer(stateReq3, {
      type: getType(actions.packages.receiveAvailablePackageSummaries) as any,
      payload: {
        response: {
          availablePackageSummaries: [availablePackageSummary1],
          nextPageToken: "3",
          categories: ["foo"],
        },
        page: 3,
      } as IReceivePackagesActionPayload,
    });
    expect(stateRec3).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: true,
      categories: ["foo"],
      items: [availablePackageSummary1, availablePackageSummary2],
    });
  });

  // TODO(agamez): check whether or not we really want to filter out duplicates. If so, add some deleted tests back

  it("two receiveAvailablePackageSummaries and then createErrorPackage", () => {
    const state1 = packageReducer(undefined, {
      type: getType(actions.packages.receiveAvailablePackageSummaries) as any,
      payload: {
        response: {
          availablePackageSummaries: [availablePackageSummary1],
          nextPageToken: "1",
          categories: ["foo"],
        },
        page: 1,
      } as IReceivePackagesActionPayload,
    });
    const state2 = packageReducer(state1, {
      type: getType(actions.packages.receiveAvailablePackageSummaries) as any,
      payload: {
        response: {
          availablePackageSummaries: [],
          nextPageToken: "2",
          categories: ["foo"],
        },
        page: 2,
      } as IReceivePackagesActionPayload,
    });
    const state3 = packageReducer(state2, {
      type: getType(actions.packages.createErrorPackage) as any,
    });
    expect(state3).toEqual({
      ...initialState,
      isFetching: false,
      categories: ["foo"],
      items: [availablePackageSummary1],
    });
  });

  it("clears errors after clearErrorPackage", () => {
    const state1 = packageReducer(undefined, {
      type: getType(actions.packages.receiveAvailablePackageSummaries) as any,
      payload: {
        response: {
          availablePackageSummaries: [availablePackageSummary1],
          nextPageToken: "5",
          categories: ["foo"],
        },
        page: 1,
      } as IReceivePackagesActionPayload,
    });
    const state2 = packageReducer(state1, {
      type: getType(actions.packages.createErrorPackage) as any,
    });
    const state3 = packageReducer(state2, {
      type: getType(actions.packages.clearErrorPackage) as any,
    });
    expect(state3).toEqual({
      ...initialState,
      isFetching: false,
      items: [availablePackageSummary1],
      categories: ["foo"],
      selected: initialState.selected,
    });
  });

  it("resetAvailablePackageSummaries resets to the initial", () => {
    const state = packageReducer(undefined, {
      type: getType(actions.packages.resetAvailablePackageSummaries) as any,
    });
    expect(state).toEqual({
      ...initialState,
    });
  });

  it("createErrorPackage resets to the initial state", () => {
    const state = packageReducer(undefined, {
      type: getType(actions.packages.createErrorPackage) as any,
    });
    expect(state).toEqual({
      ...initialState,
    });
  });
});
