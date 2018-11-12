import context from "jest-plugin-context";
import UpgradeForm from ".";
import itBehavesLike from "../../shared/specs";
import { IChartState, IChartVersion } from "../../shared/types";

const defaultProps: any = {
  fetchChartVersions: jest.fn(),
  getChartVersion: jest.fn(),
  selected: {} as IChartState["selected"],
};

describe("render", () => {
  context("when no version selected", () => {
    itBehavesLike("aLoadingComponent", { component: UpgradeForm, props: defaultProps });
  });

  context("when versions but deploying", () => {
    const version: IChartVersion = {
      id: "123",
      attributes: "lol" as any,
      relationships: "abc" as any,
    };

    itBehavesLike("aLoadingComponent", {
      component: UpgradeForm,
      props: { ...defaultProps, selected: { versions: [version], version } },
      state: { isDeploying: true },
    });
  });
});
