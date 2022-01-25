// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import messages_en from "../locales/en.json";
import { axiosWithAuth } from "./AxiosInstance";
import I18n, { II18nConfig, ISupportedLangs } from "./I18n";

const defaultI18nConfig = { locale: ISupportedLangs.en, messages: messages_en };

it("gets default i18n config", () => {
  const config: II18nConfig = I18n.getDefaultConfig();
  expect(config).toStrictEqual(defaultI18nConfig);
});

// TODO(agamez): add another lang when it is ready
it("gets existing i18n config", () => {
  const config: II18nConfig = I18n.getConfig(ISupportedLangs.en);
  const expectedConfig = { locale: ISupportedLangs.en, messages: messages_en };
  expect(config).toStrictEqual(expectedConfig);
});

it("gets default of requesting non existing i18n config", () => {
  const config: II18nConfig = I18n.getConfig("sv" as ISupportedLangs);
  const expectedConfig = defaultI18nConfig;
  expect(config).toStrictEqual(expectedConfig);
});

it("gets custom i18n config", async () => {
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: { messageId: "translation" } });
  const config: II18nConfig = await I18n.getCustomConfig(ISupportedLangs.en);
  const expected = {
    locale: ISupportedLangs.en,
    messages: { messageId: "translation" },
  };
  expect(config).toStrictEqual(expected);
});

// TODO(agamez): add another lang when it is ready
it("gets messages from the selected locale when custom i18n config fails", async () => {
  axiosWithAuth.get = jest.fn();
  const config: II18nConfig = await I18n.getCustomConfig(ISupportedLangs.en);
  const expected = { locale: ISupportedLangs.en, messages: messages_en };
  expect(config).toStrictEqual(expected);
});

it("gets default messages locale not supported and custom i18n config fails", async () => {
  axiosWithAuth.get = jest.fn();
  const config: II18nConfig = await I18n.getCustomConfig("sv" as ISupportedLangs);
  const expected = defaultI18nConfig;
  expect(config).toStrictEqual(expected);
});
