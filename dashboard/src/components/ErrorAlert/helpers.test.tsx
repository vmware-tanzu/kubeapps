import * as React from "react";

import { namespaceText } from "./helpers";

describe("namespaceText", () => {
  it("returns nothing if namespace is undefined", () => {
    expect(namespaceText()).toBe("");
  });

  it("returns the special case for all namespaces", () => {
    expect(namespaceText("_all")).toBe("all namespaces");
  });

  it("returns the element for rendering the namespace information", () => {
    expect(namespaceText("test")).toEqual(
      <span>
        the <code>test</code> namespace
      </span>,
    );
  });
});
