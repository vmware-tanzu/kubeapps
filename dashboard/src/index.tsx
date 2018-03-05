import * as React from "react";
import * as ReactDOM from "react-dom";

import Root from "./containers/Root";
import "./index.css";
// import registerServiceWorker from "./registerServiceWorker";

ReactDOM.render(<Root />, document.getElementById("root") as HTMLElement);

// Disable the service worker as it currently loads index.html from cache for
// all unknown paths. This doesn't work well with our Ingress setup because if
// you visit /kubeless/ you end up getting the Dashboard instead of the Kubeapps
// UI. We can consider re-enabling this once the Kubeless UI is integrated in
// the Dashboard and we no longer need the external link.

// registerServiceWorker();
