import * as React from "react";
import * as ReactDOM from "react-dom";
import * as Modal from "react-modal";
import { createAxiosInterceptors } from "shared/AxiosInstance";
import { axios } from "./shared/Auth";

import Root from "./containers/Root";
import "./index.css";
import store from "./store";

createAxiosInterceptors(axios, store);

ReactDOM.render(<Root />, document.getElementById("root") as HTMLElement);

// Set App Element for accessibilty
Modal.setAppElement("#root");
