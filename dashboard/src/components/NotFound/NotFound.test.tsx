// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import NotFound from "./NotFound";

it("should render the 404 page", () => {
  const wrapper = mountWrapper(defaultStore, <NotFound />);
  expect(wrapper.find(NotFound).text()).toContain("The page you are looking for can't be found");
});
