// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { act } from "@testing-library/react";
import { shallow } from "enzyme";
import { IntlProvider } from "react-intl";
import I18n, { II18nConfig } from "shared/I18n";
import Root from "./Root";

// Mocking SwaggerUI to a simple empty <div> to prevent issues with Jest
jest.mock("swagger-ui-react", () => {
  return {
    SwaggerUI: () => <div />,
  };
});

it("renders the root component", () => {
  const wrapper = shallow(<Root />);
  expect(wrapper).toExist();
});

describe("118n configuration", () => {
  it("loads the initial i18n config from getDefaultI18nConfig", async () => {
    const config: II18nConfig = { locale: "custom", messages: { messageId: "translation" } };
    const getDefaultConfig = jest.spyOn(I18n, "getDefaultConfig").mockReturnValue(config);
    act(() => {
      shallow(<Root />);
    });
    expect(getDefaultConfig).toHaveBeenCalled();
  });

  it("loads the async i18n config from getCustomI18nConfig", async () => {
    const config: II18nConfig = { locale: "custom", messages: { messageId: "translation" } };
    I18n.getCustomConfig = jest
      .fn()
      .mockReturnValue({ then: jest.fn((f: any) => f(config)), catch: jest.fn(f => f()) });
    act(() => {
      const wrapper = shallow(<Root />);
      expect(wrapper.find(IntlProvider).prop("locale")).toBe("custom");
      expect(wrapper.find(IntlProvider).prop("messages")).toStrictEqual({
        messageId: "translation",
      });
    });
  });
});
