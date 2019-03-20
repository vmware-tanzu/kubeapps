import * as React from "react";
import * as ReactDOM from "react-dom";
import * as Modal from "react-modal";
import * as Axios from "shared/AxiosInstance";
import * as Auth from "./shared/Auth";

import Root from "./containers/Root";
import "./index.css";
import store from "./store";
// import registerServiceWorker from "./registerServiceWorker";

Axios.createAxiosInterceptors(Axios.axios, store);
Axios.createAxiosInterceptorsWithAuth(Auth.axios, store);

ReactDOM.render(<Root />, document.getElementById("root") as HTMLElement);

// TODO: Look into re-enabling service worker
// registerServiceWorker();

// Set App Element for accessibilty
Modal.setAppElement("#root");
