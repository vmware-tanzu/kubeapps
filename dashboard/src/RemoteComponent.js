// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { createRemoteComponent, createRequires } from "@paciolan/remote-component";
import { resolve } from "./remote-component.config.js";

const requires = createRequires(resolve);

export const CustomComponent = createRemoteComponent({ requires });
