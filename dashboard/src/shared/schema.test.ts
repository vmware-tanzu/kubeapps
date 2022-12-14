// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { retrieveBasicFormParams, validateValuesSchema } from "./schema";
import { parseToYamlNode } from "./yamlUtils";

describe("retrieveBasicFormParams", () => {
  [
    {
      description: "should retrieve a param",
      values: "user: andres",
      schema: {
        properties: { user: { type: "string" } },
      } as any,
      result: [
        {
          type: "string",
          title: "user",
          key: "user",
          schema: { type: "string" },
          hasProperties: false,
          deployedValue: "",
          currentValue: "andres",
        },
      ],
    },
    {
      description: "should retrieve a param without default value",
      values: "user:",
      schema: {
        properties: { user: { type: "string" } },
      } as any,
      result: [
        {
          type: "string",

          title: "user",
          key: "user",
          schema: { type: "string" },
          hasProperties: false,
          deployedValue: "",
        },
      ],
    },
    {
      description: "should retrieve a param with default value in the schema",
      values: "user:",
      schema: {
        properties: { user: { type: "string", default: "michael" } },
      } as any,
      result: [
        {
          type: "string",

          default: "michael",
          title: "user",
          key: "user",
          schema: { type: "string", default: "michael" },
          hasProperties: false,
          deployedValue: "",
          defaultValue: "michael",
          currentValue: "michael",
        },
      ],
    },
    {
      description: "values prevail over default values",
      values: "user: foo",
      schema: {
        properties: { user: { type: "string", default: "bar" } },
      } as any,
      result: [
        {
          type: "string",

          default: "bar",
          title: "user",
          key: "user",
          schema: { type: "string", default: "bar" },
          hasProperties: false,
          deployedValue: "",
          defaultValue: "bar",
          currentValue: "foo",
        },
      ],
    },
    {
      description: "it should return params even if the values don't include it",
      values: "foo: bar",
      schema: {
        properties: { user: { type: "string", default: "andres" } },
      } as any,
      result: [
        {
          type: "string",

          default: "andres",
          title: "user",
          key: "user",
          schema: { type: "string", default: "andres" },
          hasProperties: false,
          deployedValue: "",
          defaultValue: "andres",
          currentValue: "andres",
        },
      ],
    },
    {
      description: "should retrieve a nested param",
      values: "credentials:\n  user: andres",
      schema: {
        properties: {
          credentials: {
            type: "object",
            properties: { user: { type: "string" } },
          },
        },
      } as any,
      result: [
        {
          type: "object",
          properties: { user: { type: "string" } },
          title: "credentials",
          key: "credentials",
          schema: { type: "object", properties: { user: { type: "string" } } },
          hasProperties: true,
          params: [
            {
              type: "string",

              title: "user",
              key: "credentials/user",
              schema: { type: "string" },
              hasProperties: false,
              deployedValue: "",
              currentValue: "andres",
            },
          ],
          deployedValue: "",
          defaultValue: "",
          currentValue: "",
        },
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
                  user: { type: "string" },
                  pass: { type: "string" },
                },
              },
            },
          },
          replicas: { type: "number" },
          service: { type: "string" },
        },
      } as any,
      result: [
        {
          type: "object",
          properties: {
            admin: {
              type: "object",
              properties: {
                user: { type: "string" },
                pass: { type: "string" },
              },
            },
          },
          title: "credentials",
          key: "credentials",
          schema: {
            type: "object",
            properties: {
              admin: {
                type: "object",
                properties: {
                  user: { type: "string" },
                  pass: { type: "string" },
                },
              },
            },
          },
          hasProperties: true,
          params: [
            {
              type: "object",
              properties: {
                user: { type: "string" },
                pass: { type: "string" },
              },
              title: "admin",
              key: "credentials/admin",
              schema: {
                type: "object",
                properties: {
                  user: { type: "string" },
                  pass: { type: "string" },
                },
              },
              hasProperties: true,
              params: [
                {
                  type: "string",

                  title: "user",
                  key: "credentials/admin/user",
                  schema: { type: "string" },
                  hasProperties: false,
                  deployedValue: "",
                  currentValue: "andres",
                },
                {
                  type: "string",

                  title: "pass",
                  key: "credentials/admin/pass",
                  schema: { type: "string" },
                  hasProperties: false,
                  deployedValue: "",
                  currentValue: "myPassword",
                },
              ],
              deployedValue: "",
              defaultValue: "",
              currentValue: "",
            },
          ],
          deployedValue: "",
          defaultValue: "",
          currentValue: "",
        },
        {
          type: "number",

          title: "replicas",
          key: "replicas",
          schema: { type: "number" },
          hasProperties: false,
          deployedValue: "",
          currentValue: 1,
        },
        {
          type: "string",
          title: "service",
          key: "service",
          schema: { type: "string" },
          hasProperties: false,
          deployedValue: "",
          currentValue: "ClusterIP",
        },
      ],
    },
    {
      description: "should retrieve a param with title and description",
      values: "blogName: myBlog",
      schema: {
        properties: {
          blogName: {
            type: "string",

            title: "Blog Name",
            description: "Title of the blog",
          },
        },
      } as any,
      result: [
        {
          type: "string",

          title: "Blog Name",
          description: "Title of the blog",
          key: "blogName",
          schema: {
            type: "string",

            title: "Blog Name",
            description: "Title of the blog",
          },
          hasProperties: false,
          deployedValue: "",
          currentValue: "myBlog",
        },
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

            properties: {
              name: { type: "string" },
              port: { type: "integer" },
            },
          },
        },
      } as any,
      result: [
        {
          type: "object",

          properties: {
            name: { type: "string" },
            port: { type: "integer" },
          },
          title: "externalDatabase",
          key: "externalDatabase",
          schema: {
            type: "object",

            properties: {
              name: { type: "string" },
              port: { type: "integer" },
            },
          },
          hasProperties: true,
          params: [
            {
              type: "string",

              title: "name",
              key: "externalDatabase/name",
              schema: { type: "string" },
              hasProperties: false,
              deployedValue: "",
              currentValue: "foo",
            },
            {
              type: "integer",

              title: "port",
              key: "externalDatabase/port",
              schema: { type: "integer" },
              hasProperties: false,
              deployedValue: "",
              currentValue: 3306,
            },
          ],
          deployedValue: "",
          defaultValue: "",
          currentValue: "",
        },
      ],
    },
    {
      description: "should retrieve a false param",
      values: "foo: false",
      schema: {
        properties: {
          foo: { type: "boolean" },
        },
      } as any,
      result: [
        {
          type: "boolean",

          title: "foo",
          key: "foo",
          schema: { type: "boolean" },
          hasProperties: false,
          deployedValue: "",
          currentValue: false,
        },
      ],
    },
    {
      description: "should retrieve a param with enum values",
      values: "databaseType: postgresql",
      schema: {
        properties: {
          databaseType: {
            type: "string",

            enum: ["mariadb", "postgresql"],
          },
        },
      } as any,
      result: [
        {
          type: "string",

          enum: ["mariadb", "postgresql"],
          title: "databaseType",
          key: "databaseType",
          schema: { type: "string", enum: ["mariadb", "postgresql"] },
          hasProperties: false,
          deployedValue: "",
          currentValue: "postgresql",
        },
      ],
    },
  ].forEach(t => {
    it(t.description, () => {
      const result = retrieveBasicFormParams(
        parseToYamlNode(t.values),
        parseToYamlNode(""),
        t.schema,
        "install",
      );
      expect(result).toMatchObject(t.result);
    });
  });
});

describe("validateValuesSchema", () => {
  [
    {
      description: "Should validate a valid object",
      values: "foo: bar\n",
      defaultValues: "foo: default\n",
      schema: {
        properties: { foo: { type: "string" } },
      },
      valid: true,
      errors: null,
    },
    {
      description: "Should error if required value not provided",
      values: "foo: bar\n",
      defaultValues: "foo: default\n",
      schema: {
        required: ["bar"],
        properties: {
          foo: { type: "string" },
          bar: { type: "string" },
        },
      },
      valid: false,
      errors: [
        {
          keyword: "required",
          instancePath: "",
          message: "must have required property 'bar'",
          params: { missingProperty: "bar" },
          schemaPath: "#/required",
        },
      ],
    },
    {
      description: "Should not error if required value has a default",
      values: "foo: bar\n",
      defaultValues: "foo: default\nbar: default\n",
      schema: {
        required: ["bar"],
        properties: {
          foo: { type: "string" },
          bar: { type: "string" },
        },
      },
      valid: true,
      errors: null,
    },
    {
      description: "Should not error if extra value not present in defaults is included",
      values: "foo: bar\nelsewhere: bang",
      defaultValues: "foo: default\nbar: default\n",
      schema: {
        required: ["bar"],
        properties: {
          foo: { type: "string" },
          bar: { type: "string" },
        },
      },
      valid: true,
      errors: null,
    },
    {
      description: "Should validate an invalid object",
      values: "foo: bar\n",
      defaultValues: "foo: 1\n",
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
      const res = validateValuesSchema(t.values, t.schema, t.defaultValues);
      expect(res.valid).toBe(t.valid);
      expect(res.errors).toEqual(t.errors);
    });
  });
});
