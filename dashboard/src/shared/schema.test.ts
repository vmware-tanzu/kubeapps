import * as YAML from "yaml";

import { getDefaultValue, getDocumentElem, retrieveBasicFormParams, setValue } from "./schema";
import { IBasicFormParam } from "./types";

describe("retrieveBasicFormParams", () => {
  [
    {
      description: "should retrieve a param",
      values: "user: andres",
      schema: { properties: { user: { type: "string", form: "username" } } },
      result: {
        username: { path: "user", value: "andres" } as IBasicFormParam,
      },
    },
    {
      description: "should retrieve a param without default value",
      values: "user:",
      schema: { properties: { user: { type: "string", form: "username" } } },
      result: {
        username: { path: "user" } as IBasicFormParam,
      },
    },
    {
      description: "should retrieve a param with default value in the schema",
      values: "user:",
      schema: { properties: { user: { type: "string", form: "username", default: "michael" } } },
      result: {
        username: { path: "user", value: "michael" } as IBasicFormParam,
      },
    },
    {
      description: "default values from values should prevail",
      values: "user: foo",
      schema: { properties: { user: { type: "string", form: "username", default: "bar" } } },
      result: {
        username: { path: "user", value: "foo" } as IBasicFormParam,
      },
    },
    {
      description: "it should return params even if the values don't include it",
      values: "foo: bar",
      schema: { properties: { user: { type: "string", form: "username", default: "andres" } } },
      result: {
        username: { path: "user", value: "andres" } as IBasicFormParam,
      },
    },
    {
      description: "should retrieve a nested param",
      values: "credentials:\n  user: andres",
      schema: {
        properties: {
          credentials: {
            type: "object",
            properties: { user: { type: "string", form: "username" } },
          },
        },
      },
      result: {
        username: {
          path: "credentials.user",
          value: "andres",
        } as IBasicFormParam,
      },
    },
    {
      description: "should retrieve several params and ignore the ones not marked",
      values: `
# Application Credentials
credentials:
  admin:
    user: andres
    pass: myPassword

# Number of Replicas
replicas: 1

# Service Type
service: ClusterIP
`,
      schema: {
        properties: {
          credentials: {
            type: "object",
            properties: {
              admin: {
                type: "object",
                properties: {
                  user: { type: "string", form: "username" },
                  pass: { type: "string", form: "password" },
                },
              },
            },
          },
          replicas: { type: "number", form: "replicas" },
          service: { type: "string" },
        },
      },
      result: {
        username: {
          path: "credentials.admin.user",
          value: "andres",
        } as IBasicFormParam,
        password: {
          path: "credentials.admin.pass",
          value: "myPassword",
        } as IBasicFormParam,
        replicas: {
          path: "replicas",
          value: 1,
        } as IBasicFormParam,
      },
    },
  ].forEach(t => {
    it(t.description, () => {
      expect(retrieveBasicFormParams(t.values, t.schema)).toEqual(t.result);
    });
  });
});

describe("getDocumentElem", () => {
  [
    {
      description: "should get the root elem",
      values: "foo: bar",
      path: "foo",
      result: "bar",
    },
    {
      description: "should return a nested value",
      values: "foo:\n  bar: foobar",
      path: "foo.bar",
      result: "foobar",
    },
    {
      description: "should return a deeply nested value",
      values: "foo:\n  bar:\n    foobar: barfoo",
      path: "foo.bar.foobar",
      result: "barfoo",
    },
    {
      description: "should return a deeply nested value as an object",
      values: "foo:\n  bar:\n    foobar: 1",
      path: "foo.bar",
      result: YAML.parseDocument("foo:\n  bar:\n    foobar: 1")
        .get("foo")
        .get("bar"),
    },
    {
      description: "should ignore an invalid value",
      values: "foo:\n  bar:\n    foobar: 1",
      path: "nope",
      result: undefined,
    },
    {
      description: "should ignore an invalid value (nested)",
      values: "foo:\n  bar:\n    foobar: 1",
      path: "not.exists",
      result: undefined,
    },
  ].forEach(t => {
    it(t.description, () => {
      const doc = YAML.parseDocument(t.values);
      expect(getDocumentElem(doc, t.path)).toEqual(t.result);
    });
  });
});

describe("getDefaultValue", () => {
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
      path: "foo.bar",
      result: "foobar",
    },
    {
      description: "should return a deeply nested value",
      values: "foo:\n  bar:\n    foobar: barfoo",
      path: "foo.bar.foobar",
      result: "barfoo",
    },
    {
      description: "should ignore an invalid value",
      values: "foo:\n  bar:\n    foobar: barfoo",
      path: "nope",
      result: undefined,
    },
    {
      description: "should ignore an invalid value (nested)",
      values: "foo:\n  bar:\n    foobar: barfoo",
      path: "not.exists",
      result: undefined,
    },
  ].forEach(t => {
    it(t.description, () => {
      expect(getDefaultValue(t.values, t.path)).toEqual(t.result);
    });
  });
});

describe("setValue", () => {
  [
    {
      description: "should set a value",
      values: "foo: bar",
      path: "foo",
      newValue: "BAR",
      result: "foo: BAR\n",
    },
    {
      description: "should set a nested value",
      values: "foo:\n  bar: foobar",
      path: "foo.bar",
      newValue: "FOOBAR",
      result: "foo:\n  bar: FOOBAR\n",
    },
    {
      description: "should set a deeply nested value",
      values: "foo:\n  bar:\n    foobar: barfoo",
      path: "foo.bar.foobar",
      newValue: "BARFOO",
      result: "foo:\n  bar:\n    foobar: BARFOO\n",
    },
    {
      description: "should add a new value",
      values: "foo: bar",
      path: "new",
      newValue: "value",
      result: "foo: bar\nnew: value\n",
    },
    {
      description: "[Not Supported] Adding a new nested value returns an error",
      values: "foo: bar",
      path: "this.new.value",
      newValue: "1",
      result: undefined,
      error: true,
    },
  ].forEach(t => {
    it(t.description, () => {
      let res: any;
      try {
        res = setValue(t.values, t.path, t.newValue);
      } catch (e) {
        expect(t.error).toBe(true);
      }
      expect(res).toEqual(t.result);
    });
  });
});
