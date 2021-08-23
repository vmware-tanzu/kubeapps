import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IBasicFormParam } from "shared/types";
import { CustomComponent } from "../../../RemoteComponent";
import CustomFormComponentLoader from "./CustomFormParam";

const param = {
  path: "enableMetrics",
  value: true,
  type: "boolean",
  customComponent: {
    className: "test",
  },
} as IBasicFormParam;

const defaultProps = {
  param,
  handleBasicFormParamChange: jest.fn(),
};

const defaultState = {
  config: { remoteComponentsUrl: "" },
};

// Mocking the window so that the injected components are imported correctly
const location = window.location;
beforeAll((): void => {
  Object.defineProperty(window, "location", {
    configurable: true,
    writable: true,
    value: { origin: "../../../../../docs/developer/examples/CustomComponent.min.js" },
  });
});
afterAll((): void => {
  // eslint-disable-next-line no-restricted-globals
  window.location = location;
});

it("should render a custom form component", () => {
  const wrapper = mountWrapper(
    getStore(defaultState),
    <CustomFormComponentLoader {...defaultProps} />,
  );
  expect(wrapper.find(CustomFormComponentLoader)).toExist();
});

it("should render the remote component", () => {
  const wrapper = mountWrapper(
    getStore(defaultState),
    <CustomFormComponentLoader {...defaultProps} />,
  );
  expect(wrapper.find(CustomComponent)).toExist();
});

it("should render the remote component with the default URL", () => {
  const wrapper = mountWrapper(
    getStore(defaultState),
    <CustomFormComponentLoader {...defaultProps} />,
  );
  expect(wrapper.find(CustomComponent)).toExist();
  expect(wrapper.find(CustomComponent).prop("url")).toContain("custom_components.js");
});

it("should render the remote component with the URL if set in the config", () => {
  const wrapper = mountWrapper(
    getStore({
      config: { remoteComponentsUrl: "www.thiswebsite.com" },
    }),
    <CustomFormComponentLoader {...defaultProps} />,
  );
  expect(wrapper.find(CustomComponent).prop("url")).toBe("www.thiswebsite.com");
});
