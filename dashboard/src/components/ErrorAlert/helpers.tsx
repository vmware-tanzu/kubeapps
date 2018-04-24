import * as React from "react";

export function namespaceText(namespace?: string) {
  if (!namespace) {
    return "";
  }
  if (namespace === "_all") {
    return "all namespaces";
  } else {
    return (
      <span>
        the <code>{namespace}</code> namespace
      </span>
    );
  }
}
