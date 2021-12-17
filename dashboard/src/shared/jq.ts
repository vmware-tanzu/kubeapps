import { IAppRepositoryFilter } from "./types";

export function toFilterRule(
  names: string,
  regex: boolean,
  exclude: boolean,
): IAppRepositoryFilter {
  let filter: IAppRepositoryFilter;
  if (regex) {
    filter = { jq: ".name | test($var)", variables: { $var: names } };
  } else {
    const namesArray = names.split(",").map(n => n.trim());
    const variables = namesArray.reduce((acc, n, i) => {
      acc[`$var${i}`] = n;
      return acc;
    }, {});
    const jq = namesArray.map((_v, i) => `.name == $var${i}`).join(" or ");
    filter = { jq, variables };
  }
  if (exclude) {
    filter.jq += " | not";
  }
  return filter;
}

export function toParams(rule: IAppRepositoryFilter) {
  const regex = rule.jq.includes("| test");
  const exclude = rule.jq.includes("| not");
  const names = Object.values(rule.variables || {}).join(", ");
  return { names, regex, exclude };
}
