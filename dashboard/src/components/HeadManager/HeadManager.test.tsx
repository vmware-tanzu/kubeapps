// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import { get } from "lodash";
import { Helmet } from "react-helmet";
import * as ReactRedux from "react-redux";
import { SupportedThemes } from "shared/Config";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import HeadManager from "./HeadManager";

let spyOnUseDispatch: jest.SpyInstance;
const defaultActions = { ...actions.config };
beforeEach(() => {
  actions.config = {
    ...defaultActions,
    getTheme: jest.fn(),
  };
  const mockDispatch = jest.fn(res => res);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  spyOnUseDispatch.mockRestore();
  actions.config = defaultActions;
});

it("should use the light theme by default", () => {
  mountWrapper(
    defaultStore,
    <HeadManager>
      <></>
    </HeadManager>,
  );
  const peek = Helmet.peek();
  expect(
    (get(peek, "linkTags") as any[]).find(l => l.href === "./clr-ui.min.css"),
  ).not.toBeUndefined();
});

it("should use the dark theme", () => {
  mountWrapper(
    getStore({ config: { theme: SupportedThemes.dark } } as Partial<IStoreState>),
    <HeadManager>
      <></>
    </HeadManager>,
  );
  const peek = Helmet.peek();
  expect(
    (get(peek, "linkTags") as any[]).find(l => l.href === "./clr-ui-dark.min.css"),
  ).not.toBeUndefined();
});
