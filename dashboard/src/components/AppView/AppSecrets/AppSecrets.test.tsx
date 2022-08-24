// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { keyForResourceRef } from "shared/ResourceRef";
import { ResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { ISecret, IStoreState } from "shared/types";
import SecretItemDatum from "../ResourceTable/ResourceItem/SecretItem/SecretItemDatum";
import AppSecrets from "./AppSecrets";

const defaultProps = {
  secretRefs: [],
};

const sampleResourceRef = {
  apiVersion: "v1",
  kind: "Secret",
  name: "foo",
  namespace: "default",
} as ResourceRef;

const secret = {
  metadata: {
    name: "foo",
  },
  type: "Opaque",
  data: {
    foo: "YmFy", // bar
    foo2: "YmFyMg==", // bar2
  } as { [s: string]: string },
} as ISecret;

it("shows a message if there are no secrets", () => {
  const wrapper = mountWrapper(defaultStore, <AppSecrets {...defaultProps} />);
  expect(wrapper.text()).toContain("The current application does not include secrets");
});

it("renders a secretItemDatum per secret", () => {
  const key = keyForResourceRef(sampleResourceRef);
  const state = getStore({
    ...initialState,
    kube: {
      ...initialState.kube,
      items: {
        ...initialState.kube.items,
        [key]: {
          isFetching: false,
          item: secret,
        },
      },
    },
  } as Partial<IStoreState>);
  const wrapper = mountWrapper(state, <AppSecrets secretRefs={[sampleResourceRef]} />);
  expect(wrapper.find(SecretItemDatum)).toHaveLength(2);
});
