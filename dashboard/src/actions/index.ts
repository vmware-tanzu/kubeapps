import { push } from "connected-react-router";
import * as apps from "./apps";
import * as auth from "./auth";
import * as packages from "./packages";
import * as config from "./config";
import * as kube from "./kube";
import * as namespace from "./namespace";
import * as operators from "./operators";
import * as repos from "./repos";

export default {
  apps,
  auth,
  packages,
  config,
  kube,
  namespace,
  operators,
  repos,
  shared: {
    pushSearchFilter: (f: string) => push(`?q=${f}`),
    pushAllNSFilter: (y: boolean) => push(`?allns=${y ? "yes" : "no"}`),
  },
};
