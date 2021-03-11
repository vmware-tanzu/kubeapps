import { toFilterRule, toParams } from "./jq";
import { IAppRepositoryFilter } from "./types";

describe("toFilterRule", () => {
  const tests = [
    {
      description: "rule with a name",
      names: "nginx",
      regex: false,
      exclude: false,
      expected: { jq: ".name == $var0", variables: { $var0: "nginx" } } as IAppRepositoryFilter,
    },
    {
      description: "rule with a several names",
      names: "nginx, jenkins",
      regex: false,
      exclude: false,
      expected: {
        jq: ".name == $var0 or .name == $var1",
        variables: { $var0: "nginx", $var1: "jenkins" },
      } as IAppRepositoryFilter,
    },
    {
      description: "rule with a regex",
      names: ".*foo.*",
      regex: true,
      exclude: false,
      expected: {
        jq: ".name | test($var)",
        variables: { $var: ".*foo.*" },
      } as IAppRepositoryFilter,
    },
    {
      description: "negated rule",
      names: "nginx",
      regex: false,
      exclude: true,
      expected: {
        jq: ".name == $var0 | not",
        variables: { $var0: "nginx" },
      } as IAppRepositoryFilter,
    },
    {
      description: "negated regex",
      names: "nginx",
      regex: true,
      exclude: true,
      expected: {
        jq: ".name | test($var) | not",
        variables: { $var: "nginx" },
      } as IAppRepositoryFilter,
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
      rule: { jq: ".name == $var0", variables: { $var0: "nginx" } } as IAppRepositoryFilter,
      expected: { names: "nginx", regex: false, exclude: false },
    },
    {
      description: "rule with a several names",
      rule: {
        jq: ".name == $var0 or .name == $var1",
        variables: { $var0: "nginx", $var1: "jenkins" },
      } as IAppRepositoryFilter,
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
      } as IAppRepositoryFilter,
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
      } as IAppRepositoryFilter,
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
      } as IAppRepositoryFilter,
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
