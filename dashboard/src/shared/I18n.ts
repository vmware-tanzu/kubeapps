// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import messages_en from "../locales/en.json";
import { axiosWithAuth } from "./AxiosInstance";
import * as url from "shared/url";

export interface II18nConfig {
  locale: ISupportedLangs | "custom";
  messages: Record<string, string>;
}

// Add here new supported languages literals
export enum ISupportedLangs {
  en = "en",
}

const messages = {};

// Load here the compiled messages for each supported language
messages[ISupportedLangs.en] = messages_en;

export default class I18n {
  public static getDefaultConfig(): II18nConfig {
    return { locale: ISupportedLangs.en, messages: messages[ISupportedLangs.en] };
  }

  public static getConfig(lang: ISupportedLangs): II18nConfig {
    if (lang && ISupportedLangs[lang]) {
      return { locale: lang, messages: messages[lang] };
    } else {
      return this.getDefaultConfig();
    }
  }

  public static async getCustomConfig(lang: ISupportedLangs) {
    try {
      const customMessages = (
        await axiosWithAuth.get<Record<string, string>>(url.api.custom_locale)
      ).data;
      if (Object.keys(customMessages).length === 0) {
        throw new Error("Empty custom locale");
      }
      return { locale: lang, messages: customMessages };
    } catch (e: any) {
      if (lang && ISupportedLangs[lang]) {
        return this.getConfig(lang);
      } else {
        return this.getDefaultConfig();
      }
    }
  }
}
