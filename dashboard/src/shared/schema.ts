import { IBasicFormParam } from "./types";

export function retrieveBasicFormParams(schema: any) {
  // TBD
  return {
    username: {
      name: "username",
      path: "wordpressUsername",
      value: "user",
    } as IBasicFormParam,
  };
}

export function setValue(values: string, path: string, newValue: any) {
  // TBD
  return values;
}
