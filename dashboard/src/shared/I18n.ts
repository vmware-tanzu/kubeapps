import messages_en from "../locales/en.json";
import { axiosWithAuth } from "./AxiosInstance";

export interface II18nConfig {
  locale: string;
  messages: Record<string, string>;
}

export enum ISupportedLangs {
  en = "en",
  custom = "custom",
}

const messages = {};
messages[ISupportedLangs.en] = messages_en;

export function getDefaulI18nConfig(): II18nConfig {
  return { locale: ISupportedLangs.en, messages: messages[ISupportedLangs.en] };
}

export function getI18nConfig(lang: ISupportedLangs): II18nConfig {
  if (lang && ISupportedLangs[lang]) {
    return { locale: lang, messages: messages[lang] };
  } else {
    return getDefaulI18nConfig();
  }
}

export async function getCustomI18nConfig(lang: ISupportedLangs) {
  try {
    messages[ISupportedLangs.custom] = (
      await axiosWithAuth.get<Record<string, string>>("custom_locale.json")
    ).data;
    if (Object.keys(messages[ISupportedLangs.custom]).length === 0) {
      throw new Error("Empty custom locale");
    }
    return { locale: lang, messages: messages[ISupportedLangs.custom] };
  } catch (err) {
    if (lang && ISupportedLangs[lang]) {
      return getI18nConfig(lang);
    } else {
      return getDefaulI18nConfig();
    }
  }
}
