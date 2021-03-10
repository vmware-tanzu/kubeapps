import AceEditor from "react-ace";
import { SupportedThemes } from "shared/Config";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import AdvancedDeploymentForm from "./AdvancedDeploymentForm";

const defaultProps = {
  handleValuesChange: jest.fn(),
};

it("includes values", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <AdvancedDeploymentForm {...defaultProps} appValues="foo: bar" />,
  );
  expect(wrapper.find(AceEditor).prop("value")).toBe("foo: bar");
});

it("sets light theme by default", () => {
  const wrapper = mountWrapper(defaultStore, <AdvancedDeploymentForm {...defaultProps} />);
  expect(wrapper.find(AceEditor).prop("theme")).toBe("xcode");
});

it("changes theme", () => {
  const wrapper = mountWrapper(
    getStore({ config: { theme: SupportedThemes.dark } }),
    <AdvancedDeploymentForm {...defaultProps} />,
  );
  expect(wrapper.find(AceEditor).prop("theme")).toBe("solarized_dark");
});
