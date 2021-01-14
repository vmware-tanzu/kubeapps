import * as React from "react";
import * as ReactDOM from "react-dom";
import Modal from "react-modal";
import { addAuthHeaders, addErrorHandling, axios, axiosWithAuth } from "shared/AxiosInstance";

import Root from "./containers/Root";
import "./index.css";
import store from "./store";
// import registerServiceWorker from "./registerServiceWorker";

// Now that the store has been initialized, initialize axios instances
// One axios instance will be used for services that requires auth (those that use the K8s API)
// and the other for services that don't (like the assetsvc)
addErrorHandling(axios, store);
addErrorHandling(axiosWithAuth, store);
addAuthHeaders(axiosWithAuth);

ReactDOM.render(<Root />, document.getElementById("root") as HTMLElement);

// Allow to load custom styling.
// The dashboard webserver will return this style file.
const link = document.createElement("link");
link.href = "/custom_style.css";
link.type = "text/css";
link.rel = "stylesheet";
document.getElementsByTagName("head")[0].appendChild(link);

// TODO: Look into re-enabling service worker
// registerServiceWorker();

// Set App Element for accessibilty
Modal.setAppElement("#root");
