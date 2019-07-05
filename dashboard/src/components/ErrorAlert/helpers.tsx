import * as React from "react";
import { definedNamespaces } from "../../shared/Namespace";

export function namespaceText(namespace?: string) {
  if (!namespace) {
    return "";
  }
  if (namespace === definedNamespaces.all) {
    return "all namespaces";
  } else {
    return (
      <span>
        the <code>{namespace}</code> namespace
      </span>
    );
  }
}
