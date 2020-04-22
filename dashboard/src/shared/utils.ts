export function escapeRegExp(str: string) {
  return str.replace(/[\-\[\]\/\{\}\(\)\*\+\?\.\\\^\$\|]/g, "\\$&");
}

export function getValueFromEvent(e: React.FormEvent<HTMLInputElement>) {
  let value: any = e.currentTarget.value;
  switch (e.currentTarget.type) {
    case "checkbox":
      // value is a boolean
      value = e.currentTarget.checked;
      break;
    case "number":
      // value is a number
      value = e.currentTarget.valueAsNumber;
      break;
  }
  return value;
}

export async function wait(ms: number = 1) {
  await new Promise(resolve => setTimeout(() => resolve(), ms));
}
