export function escapeRegExp(str: string) {
  return str.replace(/[-[\]/{}()*+?.\\^$|]/g, "\\$&");
}

export function getValueFromEvent(
  e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>,
) {
  let value: any = e.currentTarget.value;
  switch (e.currentTarget.type) {
    case "checkbox":
      // value is a boolean
      value = value === "true";
      break;
    case "number":
      // value is a number
      value = parseInt(value, 10);
      break;
  }
  return value;
}

export async function wait(ms: number = 1) {
  await new Promise(resolve => setTimeout(() => resolve(), ms));
}

// 3 lines description max
const MAX_DESC_LENGTH = 90;

export function trimDescription(desc: string): string {
  if (desc.length > MAX_DESC_LENGTH) {
    // Trim to the last word under the max length
    return desc.substr(0, desc.lastIndexOf(" ", MAX_DESC_LENGTH)).concat("...");
  }
  return desc;
}
