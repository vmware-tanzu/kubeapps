import { mount } from "enzyme";
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
  const wrapper = mount(<CustomFormComponentLoader {...defaultProps} />);
  expect(wrapper.find(CustomFormComponentLoader)).toExist();
});

it("should render the remote component", () => {
  const wrapper = mount(<CustomFormComponentLoader {...defaultProps} />);
  expect(wrapper.find(CustomComponent)).toExist();
});
