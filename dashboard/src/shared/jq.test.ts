// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { toFilterRule, toParams } from "./jq";
import { IPkgRepositoryFilter } from "./types";

describe("toFilterRule", () => {
  const tests = [
    {
      description: "rule with a name",
      names: "nginx",
      regex: false,
      exclude: false,
      expected: { jq: ".name == $var0", variables: { $var0: "nginx" } } as IPkgRepositoryFilter,
    },
    {
      description: "rule with a several names",
      names: "nginx, jenkins",
      regex: false,
      exclude: false,
      expected: {
        jq: ".name == $var0 or .name == $var1",
        variables: { $var0: "nginx", $var1: "jenkins" },
      } as IPkgRepositoryFilter,
    },
    {
      description: "rule with a regex",
      names: ".*foo.*",
      regex: true,
      exclude: false,
      expected: {
        jq: ".name | test($var)",
        variables: { $var: ".*foo.*" },
      } as IPkgRepositoryFilter,
    },
    {
      description: "negated rule",
      names: "nginx",
      regex: false,
      exclude: true,
      expected: {
        jq: ".name == $var0 | not",
        variables: { $var0: "nginx" },
      } as IPkgRepositoryFilter,
    },
    {
      description: "negated regex",
      names: "nginx",
      regex: true,
      exclude: true,
      expected: {
        jq: ".name | test($var) | not",
        variables: { $var: "nginx" },
      } as IPkgRepositoryFilter,
    },
  ];
  tests.forEach(t => {
    it(t.description, () => {
      expect(toFilterRule(t.names, t.regex, t.exclude)).toEqual(t.expected);
    });
  });
});

describe("toParams", () => {
  const tests = [
    {
      description: "rule with a name",
      rule: { jq: ".name == $var0", variables: { $var0: "nginx" } } as IPkgRepositoryFilter,
      expected: { names: "nginx", regex: false, exclude: false },
    },
    {
      description: "rule with a several names",
      rule: {
        jq: ".name == $var0 or .name == $var1",
        variables: { $var0: "nginx", $var1: "jenkins" },
      } as IPkgRepositoryFilter,
      expected: {
        names: "nginx, jenkins",
        regex: false,
        exclude: false,
      },
    },
    {
      description: "rule with a regex",
      rule: {
        jq: ".name | test($var)",
        variables: { $var: ".*foo.*" },
      } as IPkgRepositoryFilter,
      expected: {
        names: ".*foo.*",
        regex: true,
        exclude: false,
      },
    },
    {
      description: "negated rule",
      rule: {
        jq: ".name == $var0 | not",
        variables: { $var0: "nginx" },
      } as IPkgRepositoryFilter,
      expected: {
        names: "nginx",
        regex: false,
        exclude: true,
      },
    },
    {
      description: "negated regex",
      rule: {
        jq: ".name | test($var) | not",
        variables: { $var: "nginx" },
      } as IPkgRepositoryFilter,
      expected: {
        names: "nginx",
        regex: true,
        exclude: true,
      },
    },
  ];
  tests.forEach(t => {
    it(t.description, () => {
      expect(toParams(t.rule)).toEqual(t.expected);
    });
  });
});
