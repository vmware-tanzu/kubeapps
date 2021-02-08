import messages_en from "../locales/en.json";
import { axiosWithAuth } from "./AxiosInstance";
import {
  getCustomI18nConfig,
  getDefaulI18nConfig,
  getI18nConfig,
  II18nConfig,
  ISupportedLangs,
} from "./I18n";

const defaultI18nConfig = { locale: ISupportedLangs.en, messages: messages_en };

it("gets default i18n config", () => {
  const config: II18nConfig = getDefaulI18nConfig();
  const expectedConfig: II18nConfig = defaultI18nConfig;
  expect(config).toStrictEqual(expectedConfig);
});

// TODO(agamez): add another lang when it is ready
it("gets existing i18n config", () => {
  const config: II18nConfig = getI18nConfig(ISupportedLangs.en);
  const expectedConfig: II18nConfig = { locale: ISupportedLangs.en, messages: messages_en };
  expect(config).toStrictEqual(expectedConfig);
});

it("gets default of requesting non existing i18n config", () => {
  const config: II18nConfig = getI18nConfig("sv" as ISupportedLangs);
  const expectedConfig: II18nConfig = defaultI18nConfig;
  expect(config).toStrictEqual(expectedConfig);
});

it("gets custom i18n config", async () => {
  const data: Record<string, string> = { messageId: "translation" };
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: data });
  const config: II18nConfig = await getCustomI18nConfig(ISupportedLangs.en);
  const expected: II18nConfig = { locale: ISupportedLangs.en, messages: data };
  expect(config).toStrictEqual(expected);
});

// TODO(agamez): add another lang when it is ready
it("gets messages from the selected locale when custom i18n config fails", async () => {
  axiosWithAuth.get = jest.fn();
  const config: II18nConfig = await getCustomI18nConfig(ISupportedLangs.en);
  const expected: II18nConfig = { locale: ISupportedLangs.en, messages: messages_en };
  expect(config).toStrictEqual(expected);
});

it("gets default messages locale not supported and custom i18n config fails", async () => {
  axiosWithAuth.get = jest.fn();
  const config: II18nConfig = await getCustomI18nConfig("sv" as ISupportedLangs);
  const expected: II18nConfig = defaultI18nConfig;
  expect(config).toStrictEqual(expected);
});
