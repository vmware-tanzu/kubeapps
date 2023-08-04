// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import * as installedpackages from "./installedpackages";
import * as auth from "./auth";
import * as availablepackages from "./availablepackages";
import * as config from "./config";
import * as kube from "./kube";
import * as namespace from "./namespace";
import * as operators from "./operators";
import * as repos from "./repos";

export default {
  installedpackages,
  auth,
  availablepackages,
  config,
  kube,
  namespace,
  operators,
  repos,
};
