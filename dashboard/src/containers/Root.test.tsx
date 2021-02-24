import ThemeSelector, { SupportedThemes } from "components/ThemeSelector/ThemeSelector";
import { shallow } from "enzyme";
import { act } from "react-dom/test-utils";
import { IntlProvider } from "react-intl";
import I18n, { II18nConfig } from "shared/I18n";
import Root from "./Root";

it("renders the root component", () => {
  const wrapper = shallow(<Root />);
  expect(wrapper).toExist();
});

describe("118n configuration", () => {
  it("loads the initial i18n config from getDefaulI18nConfig", async () => {
    const config: II18nConfig = { locale: "custom", messages: { messageId: "translation" } };
    const getDefaultConfig = jest.spyOn(I18n, "getDefaultConfig").mockReturnValue(config);
    act(() => {
      shallow(<Root />);
    });
    expect(getDefaultConfig).toHaveBeenCalled();
  });

  it("loads the async i18n config from getCustomI18nConfig", async () => {
    const config: II18nConfig = { locale: "custom", messages: { messageId: "translation" } };
    I18n.getCustomConfig = jest.fn().mockReturnValue({ then: jest.fn((f: any) => f(config)) });
    act(() => {
      const wrapper = shallow(<Root />);
      expect(wrapper.find(IntlProvider).prop("locale")).toBe("custom");
      expect(wrapper.find(IntlProvider).prop("messages")).toStrictEqual({
        messageId: "translation",
      });
    });
  });
});

describe("css initial configuration", () => {
  it("load the default 'light' theme", () => {
    const wrapper = shallow(<Root />);
    expect(wrapper.find(ThemeSelector).prop("theme")).toBe(SupportedThemes.light);
  });

  it("load the stored theme", () => {
    localStorage.setItem("theme", SupportedThemes.dark);
    const wrapper = shallow(<Root />);
    expect(wrapper.find(ThemeSelector).prop("theme")).toBe(SupportedThemes.dark);
  });
});
