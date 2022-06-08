// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import ReactDOM from "react-dom";
import { addAuthHeaders, addErrorHandling, axios, axiosWithAuth } from "shared/AxiosInstance";
import Root from "./containers/Root";
import "./index.css";
import store from "./store";

// import registerServiceWorker from "./registerServiceWorker";

// Now that the store has been initialized, initialize axios instances
// One axios instance will be used for services that requires auth (those that use the K8s API)
// and the other for services that don't
addErrorHandling(axios, store);
addErrorHandling(axiosWithAuth, store);
addAuthHeaders(axiosWithAuth);

ReactDOM.render(<Root />, document.getElementById("root") as HTMLElement);

// TODO: Look into re-enabling service worker
// registerServiceWorker();
