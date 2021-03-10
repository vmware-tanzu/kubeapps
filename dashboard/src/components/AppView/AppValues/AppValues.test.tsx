import AceEditor from "react-ace";
import { SupportedThemes } from "shared/Config";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import AppValues from "./AppValues";

it("includes values", () => {
  const wrapper = mountWrapper(defaultStore, <AppValues values="foo: bar" />);
  expect(wrapper.find(AceEditor).prop("value")).toBe("foo: bar");
});

it("adds a default message if no values are given", () => {
  const wrapper = mountWrapper(defaultStore, <AppValues values="" />);
  expect(wrapper.find(AceEditor)).not.toExist();
  expect(wrapper).toIncludeText(
    "The current application was installed without specifying any values",
  );
});

it("sets light theme by default", () => {
  const wrapper = mountWrapper(defaultStore, <AppValues values="foo: bar" />);
  expect(wrapper.find(AceEditor).prop("theme")).toBe("xcode");
});

it("changes theme", () => {
  const wrapper = mountWrapper(
    getStore({ config: { theme: SupportedThemes.dark } }),
    <AppValues values="foo: bar" />,
  );
  expect(wrapper.find(AceEditor).prop("theme")).toBe("solarized_dark");
});
