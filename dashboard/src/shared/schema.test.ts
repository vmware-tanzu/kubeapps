// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { deleteValue, getValue, retrieveBasicFormParams, setValue, validate } from "./schema";
import { IBasicFormParam } from "./types";

describe("retrieveBasicFormParams", () => {
  [
    {
      description: "should retrieve a param",
      values: "user: andres",
      schema: {
        properties: { user: { type: "string", form: true } },
      } as any,
      result: [
        {
          path: "user",
          value: "andres",
        } as IBasicFormParam,
      ],
    },
    {
      description: "should retrieve a param without default value",
      values: "user:",
      schema: {
        properties: { user: { type: "string", form: true } },
      } as any,
      result: [
        {
          path: "user",
        } as IBasicFormParam,
      ],
    },
    {
      description: "should retrieve a param with default value in the schema",
      values: "user:",
      schema: {
        properties: { user: { type: "string", form: true, default: "michael" } },
      } as any,
      result: [
        {
          path: "user",
          value: "michael",
        } as IBasicFormParam,
      ],
    },
    {
      description: "values prevail over default values",
      values: "user: foo",
      schema: {
        properties: { user: { type: "string", form: true, default: "bar" } },
      } as any,
      result: [
        {
          path: "user",
          value: "foo",
        } as IBasicFormParam,
      ],
    },
    {
      description: "it should return params even if the values don't include it",
      values: "foo: bar",
      schema: {
        properties: { user: { type: "string", form: true, default: "andres" } },
      } as any,
      result: [
        {
          path: "user",
          value: "andres",
        } as IBasicFormParam,
      ],
    },
    {
      description: "should retrieve a nested param",
      values: "credentials:\n  user: andres",
      schema: {
        properties: {
          credentials: {
            type: "object",
            properties: { user: { type: "string", form: true } },
          },
        },
      } as any,
      result: [
        {
          path: "credentials/user",
          value: "andres",
        } as IBasicFormParam,
      ],
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
                  user: { type: "string", form: true },
                  pass: { type: "string", form: true },
                },
              },
            },
          },
          replicas: { type: "number", form: true },
          service: { type: "string" },
        },
      } as any,
      result: [
        {
          path: "credentials/admin/user",
          value: "andres",
        } as IBasicFormParam,
        {
          path: "credentials/admin/pass",
          value: "myPassword",
        } as IBasicFormParam,
        {
          path: "replicas",
          value: 1,
        } as IBasicFormParam,
      ],
    },
    {
      description: "should retrieve a param with title and description",
      values: "blogName: myBlog",
      schema: {
        properties: {
          blogName: {
            type: "string",
            form: true,
            title: "Blog Name",
            description: "Title of the blog",
          },
        },
      } as any,
      result: [
        {
          path: "blogName",
          type: "string",
          value: "myBlog",
          title: "Blog Name",
          description: "Title of the blog",
        } as IBasicFormParam,
      ],
    },
    {
      description: "should retrieve a param with children params",
      values: `
externalDatabase:
  name: "foo"
  port: 3306
`,
      schema: {
        properties: {
          externalDatabase: {
            type: "object",
            form: true,
            properties: {
              name: { type: "string", form: true },
              port: { type: "integer", form: true },
            },
          },
        },
      } as any,
      result: [
        {
          path: "externalDatabase",
          type: "object",
          children: [
            {
              path: "externalDatabase/name",
              type: "string",
            },
            {
              path: "externalDatabase/port",
              type: "integer",
            },
          ],
        } as IBasicFormParam,
      ],
    },
    {
      description: "should retrieve a false param",
      values: "foo: false",
      schema: {
        properties: {
          foo: { type: "boolean", form: true },
        },
      } as any,
      result: [{ path: "foo", type: "boolean", value: false } as IBasicFormParam],
    },
    {
      description: "should retrieve a param with enum values",
      values: "databaseType: postgresql",
      schema: {
        properties: {
          databaseType: {
            type: "string",
            form: true,
            enum: ["mariadb", "postgresql"],
          },
        },
      } as any,
      result: [
        {
          path: "databaseType",
          type: "string",
          value: "postgresql",
          enum: ["mariadb", "postgresql"],
        } as IBasicFormParam,
      ],
    },
  ].forEach(t => {
    it(t.description, () => {
      expect(retrieveBasicFormParams(t.values, t.schema)).toMatchObject(t.result);
    });
  });
});

describe("getValue", () => {
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
      expect(getValue(t.values, t.path, t.default)).toEqual(t.result);
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

describe("validate", () => {
  [
    {
      description: "Should validate a valid object",
      values: "foo: bar\n",
      schema: {
        properties: { foo: { type: "string" } },
      },
      valid: true,
      errors: null,
    },
    {
      description: "Should validate an invalid object",
      values: "foo: bar\n",
      schema: { properties: { foo: { type: "integer" } } },
      valid: false,
      errors: [
        {
          keyword: "type",
          instancePath: "/foo",
          schemaPath: "#/properties/foo/type",
          params: {
            type: "integer",
          },
          message: "must be integer",
        },
      ],
    },
  ].forEach(t => {
    it(t.description, () => {
      const res = validate(t.values, t.schema);
      expect(res.valid).toBe(t.valid);
      expect(res.errors).toEqual(t.errors);
    });
  });
});
