import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";
import itBehavesLike from "../../shared/specs";
import { IChartState, IChartVersion } from "../../shared/types";
import UpgradeForm from "./UpgradeForm";

const version: IChartVersion = {
  id: "123",
  attributes: "lol" as any,
  relationships: "abc" as any,
};

const defaultProps: any = {
  fetchChartVersions: jest.fn(),
  getChartVersion: jest.fn(),
  selected: {} as IChartState["selected"],
  goBack: jest.fn(),
  upgradeApp: jest.fn(),
};

describe("render", () => {
  context("when no version selected", () => {
    itBehavesLike("aLoadingComponent", { component: UpgradeForm, props: defaultProps });
  });

  context("when versions but deploying", () => {
    itBehavesLike("aLoadingComponent", {
      component: UpgradeForm,
      props: { ...defaultProps, selected: { versions: [version], version } },
      state: { isDeploying: true },
    });
  });
});

it("goes back when clicking in the Back button", () => {
  const goBack = jest.fn();
  const upgradeApp = jest.fn();
  const selected = {
    version,
    versions: [version],
  } as IChartState["selected"];
  const wrapper = shallow(
    <UpgradeForm {...defaultProps} goBack={goBack} upgradeApp={upgradeApp} selected={selected} />,
  );
  const backButton = wrapper.find(".button").filterWhere(i => i.text() === "Back");
  expect(backButton).toExist();
  // Avoid empty or submit type
  expect(backButton.prop("type")).toBe("button");
  backButton.simulate("click");
  expect(goBack).toBeCalled();
  expect(upgradeApp).not.toBeCalled();
});
