import { push } from "react-router-redux";

import * as apps from "./apps";
import * as auth from "./auth";
import * as catalog from "./catalog";
import * as charts from "./charts";
import * as config from "./config";
import * as functions from "./functions";
import * as namespace from "./namespace";
import * as repos from "./repos";

export default {
  apps,
  auth,
  catalog,
  charts,
  config,
  functions,
  namespace,
  repos,
  shared: {
    pushSearchFilter: (f: string) => push(`?q=${f}`),
  },
};
