import { get } from "lodash";
import { IResource } from "shared/types";

export const OtherResourceColumns = [
  {
    accessor: "name",
    Header: "NAME",
    getter: (target: IResource) => get(target, "metadata.name"),
  },
  {
    accessor: "kind",
    Header: "KIND",
    getter: (target: IResource) => get(target, "kind"),
  },
];
