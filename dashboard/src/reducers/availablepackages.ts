// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { JSONSchemaType } from "ajv";
import { uniqBy } from "lodash";
import { IPackageState } from "shared/types";
import { getType } from "typesafe-actions";
import actions from "../actions";
import { PackagesAction } from "../actions/availablepackages";
import { NamespaceAction } from "../actions/namespace";

export const initialState: IPackageState = {
  isFetching: false,
  hasFinishedFetching: false,
  items: [],
  categories: [],
  selected: {
    versions: [],
  },
  size: 20,
};

const selectedPackageReducer = (
  state: IPackageState["selected"],
  action: PackagesAction | NamespaceAction,
): IPackageState["selected"] => {
  switch (action.type) {
    case getType(actions.availablepackages.receiveSelectedAvailablePackageDetail):
      return {
        ...state,
        error: undefined,
        readmeError: undefined,
        availablePackageDetail: action.payload.selectedPackage,
        pkgVersion: action.payload.selectedPackage.version?.pkgVersion,
        appVersion: action.payload.selectedPackage.version?.appVersion,
        readme: action.payload.selectedPackage.readme,
        values: action.payload.selectedPackage.defaultValues,
        schema:
          action.payload.selectedPackage.valuesSchema !== ""
            ? (JSON.parse(action.payload.selectedPackage.valuesSchema) as JSONSchemaType<any>)
            : ({} as JSONSchemaType<any>),
      };
    case getType(actions.availablepackages.receiveSelectedAvailablePackageVersions):
      return {
        ...state,
        error: undefined,
        versions: action.payload.packageAppVersions,
      };
    case getType(actions.availablepackages.createErrorPackage):
      return { ...state, error: action.payload };
    case getType(actions.availablepackages.clearErrorPackage):
      return { ...state, error: undefined };
    case getType(actions.availablepackages.resetSelectedAvailablePackageDetail):
      return initialState.selected;
    default:
  }
  return state;
};

const packageReducer = (
  state: IPackageState = initialState,
  action: PackagesAction | NamespaceAction,
): IPackageState => {
  switch (action.type) {
    case getType(actions.availablepackages.requestAvailablePackageSummaries):
      return { ...state, isFetching: true };
    case getType(actions.availablepackages.requestSelectedAvailablePackageVersions):
      return { ...state, isFetching: true };
    case getType(actions.availablepackages.receiveAvailablePackageSummaries): {
      const isLastPage =
        action.payload.page >= parseInt(action.payload.response.nextPageToken) ||
        action.payload.response.nextPageToken === "";
      return {
        ...state,
        isFetching: false,
        hasFinishedFetching: isLastPage,
        categories: action.payload.response.categories,
        items: uniqBy(
          [...state.items, ...action.payload.response.availablePackageSummaries],
          "availablePackageRef.identifier",
        ),
      };
    }
    case getType(actions.availablepackages.receiveSelectedAvailablePackageVersions):
      return {
        ...state,
        isFetching: false,
        selected: selectedPackageReducer(state.selected, action),
      };
    case getType(actions.availablepackages.requestSelectedAvailablePackageDetail):
      return {
        ...state,
        isFetching: true,
        selected: selectedPackageReducer(state.selected, action),
      };
    case getType(actions.availablepackages.receiveSelectedAvailablePackageDetail):
      return {
        ...state,
        isFetching: false,
        selected: selectedPackageReducer(state.selected, action),
      };
    case getType(actions.availablepackages.resetAvailablePackageSummaries):
      return {
        ...state,
        hasFinishedFetching: false,
        items: [],
      };
    case getType(actions.availablepackages.createErrorPackage):
      return {
        ...state,
        isFetching: false,
        hasFinishedFetching: false,
        items: state.items,
        selected: selectedPackageReducer(state.selected, action),
      };
    case getType(actions.availablepackages.resetSelectedAvailablePackageDetail):
    case getType(actions.availablepackages.clearErrorPackage):
      return {
        ...state,
        selected: selectedPackageReducer(state.selected, action),
      };
  }
  return state;
};

export default packageReducer;
