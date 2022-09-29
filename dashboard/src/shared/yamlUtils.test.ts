// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  deleteValue,
  getPathValueInYamlNodeWithDefault,
  parseToYamlNode,
  setValue,
} from "./yamlUtils";

describe("getPathValueInYamlNodeWithDefault", () => {
  [
    {
      description: "should return a value",
      values: "foo: bar",
      path: "foo",
      result: "bar",
    },
    {
      description: "should return a nested value",
      values: "foo:\n  bar: foobar",
      path: "foo/bar",
      result: "foobar",
    },
    {
      description: "should return a deeply nested value",
      values: "foo:\n  bar:\n    foobar: barfoo",
      path: "foo/bar/foobar",
      result: "barfoo",
    },
    {
      description: "should ignore an invalid path",
      values: "foo:\n  bar:\n    foobar: barfoo",
      path: "nope",
      result: undefined,
    },
    {
      description: "should ignore an invalid path (nested)",
      values: "foo:\n  bar:\n    foobar: barfoo",
      path: "not/exists",
      result: undefined,
    },
    {
      description: "should return the default value if the path is not valid",
      values: "foo: bar",
      path: "foobar",
      default: '"BAR"',
      result: '"BAR"',
    },
    {
      description: "should return a value with slashes in the key",
      values: "foo/bar: value",
      path: "foo~1bar",
      result: "value",
    },
    {
      description: "should return a value with slashes and dots in the key",
      values: "kubernetes.io/ingress.class: nginx",
      path: "kubernetes.io~1ingress.class",
      result: "nginx",
    },
  ].forEach(t => {
    it(t.description, () => {
      expect(
        getPathValueInYamlNodeWithDefault(parseToYamlNode(t.values), t.path, t.default),
      ).toEqual(t.result);
    });
  });
});

describe("setValue", () => {
  [
    {
      description: "should set a value",
      values: 'foo: "bar"',
      path: "foo",
      newValue: "BAR",
      result: 'foo: "BAR"\n',
    },
    {
      description: "should set a value preserving the existing scalar quotation (simple)",
      values: "foo: 'bar'",
      path: "foo",
      newValue: "BAR",
      result: "foo: 'BAR'\n",
    },
    {
      description: "should set a value preserving the existing scalar quotation (double)",
      values: 'foo: "bar"',
      path: "foo",
      newValue: "BAR",
      result: 'foo: "BAR"\n',
    },
    {
      description: "should set a value preserving the existing scalar quotation (none)",
      values: "foo: bar",
      path: "foo",
      newValue: "BAR",
      result: "foo: BAR\n",
    },
    {
      description: "should set a nested value",
      values: 'foo:\n  bar: "foobar"',
      path: "foo/bar",
      newValue: "FOOBAR",
      result: 'foo:\n  bar: "FOOBAR"\n',
    },
    {
      description: "should set a deeply nested value",
      values: 'foo:\n  bar:\n    foobar: "barfoo"',
      path: "foo/bar/foobar",
      newValue: "BARFOO",
      result: 'foo:\n  bar:\n    foobar: "BARFOO"\n',
    },
    {
      description: "should add a new value",
      values: "foo: bar",
      path: "new",
      newValue: "value",
      result: 'foo: bar\nnew: "value"\n',
    },
    {
      description: "should add a new nested value",
      values: "foo: bar",
      path: "this/new",
      newValue: 1,
      result: "foo: bar\nthis:\n  new: 1\n",
      error: false,
    },
    {
      description: "should add a new deeply nested value",
      values: "foo: bar",
      path: "this/new/value",
      newValue: 1,
      result: "foo: bar\nthis:\n  new:\n    value: 1\n",
      error: false,
    },
    {
      description: "Adding a value for a path partially defined (null)",
      values: "foo: bar\nthis:\n",
      path: "this/new/value",
      newValue: 1,
      result: "foo: bar\nthis:\n  new:\n    value: 1\n",
      error: false,
    },
    {
      description: "Adding a value for a path partially defined (object)",
      values: "foo: bar\nthis: {}\n",
      path: "this/new/value",
      newValue: 1,
      result: "foo: bar\nthis: { new: { value: 1 } }\n",
      error: false,
    },
    {
      description: "Adding a value in an empty doc",
      values: "",
      path: "foo",
      newValue: "bar",
      result: 'foo: "bar"\n',
      error: false,
    },
    {
      description: "should add a value with slashes in the key",
      values: 'foo/bar: "test"',
      path: "foo~1bar",
      newValue: "value",
      result: 'foo/bar: "value"\n',
    },
    {
      description: "should add a value with slashes and dots in the key",
      values: 'kubernetes.io/ingress.class: "default"',
      path: "kubernetes.io~1ingress.class",
      newValue: "nginx",
      result: 'kubernetes.io/ingress.class: "nginx"\n',
    },
  ].forEach(t => {
    it(t.description, () => {
      if (t.error) {
        expect(() => setValue(t.values, t.path, t.newValue)).toThrow();
      } else {
        expect(setValue(t.values, t.path, t.newValue)).toEqual(t.result);
      }
    });
  });
});

describe("deleteValue", () => {
  [
    {
      description: "should delete a value",
      values: "foo: bar\nbar: foo\n",
      path: "bar",
      result: "foo: bar\n",
    },
    {
      description: "should delete a value from an array",
      values: `foo:
  - bar
  - foobar
`,
      path: "foo/0",
      result: `foo:
  - foobar
`,
    },
    {
      description: "should leave the document empty",
      values: "foo: bar",
      path: "foo",
      result: "\n",
    },
    {
      description: "noop when trying to delete a missing property",
      values: "foo: bar\nbar: foo\n",
      path: "var",
      result: "foo: bar\nbar: foo\n",
    },
  ].forEach(t => {
    it(t.description, () => {
      expect(deleteValue(t.values, t.path)).toEqual(t.result);
    });
  });
});
