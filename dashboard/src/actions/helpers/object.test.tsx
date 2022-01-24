// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import * as object from "./object";

describe("isEmptyDeep", () => {
  [
    { object: {}, isEmpty: true },
    { object: { a: 1 }, isEmpty: false },
    { object: { a: { b: "" } }, isEmpty: true },
    { object: { a: { c: { d: "" } } }, isEmpty: true },
    { object: { a: { c: { d: "a" } } }, isEmpty: false },
    { object: { a: { c: { d: [] } } }, isEmpty: true },
    { object: { a: { c: { d: {} } } }, isEmpty: true },
    { object: { a: { c: { d: "a" }, d: "" } }, isEmpty: false },
  ].forEach(t => {
    it(`${JSON.stringify(t.object)} should be ${t.isEmpty ? "empty" : "not empty"}`, () => {
      expect(object.isEmptyDeep(t.object)).toBe(t.isEmpty);
    });
  });
});

describe("removeEmptyField", () => {
  [
    { in: {}, out: {} },
    { in: { a: 1 }, out: { a: 1 } },
    { in: { a: { b: "" } }, out: {} },
    { in: { a: { b: "" }, c: 1 }, out: { c: 1 } },
    { in: { a: { c: { d: "" } } }, out: {} },
    { in: { a: { c: { d: "a" } } }, out: { a: { c: { d: "a" } } } },
    { in: { a: { c: { d: [] } } }, out: {} },
    { in: { a: { c: { d: {} } } }, out: {} },
  ].forEach(t => {
    it(`${JSON.stringify(t.in)} should be simplified to ${JSON.stringify(t.out)}`, () => {
      expect(object.removeEmptyFields(t.in)).toEqual(t.out);
    });
  });
});
