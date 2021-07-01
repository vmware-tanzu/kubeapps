import { CdsButton } from "@cds/react/button";
import actions from "actions";
import { hapi } from "../../../shared/hapi/release";

import * as ReactRedux from "react-redux";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import StatusAwareButton from "./StatusAwareButton";
import ReactTooltip from "react-tooltip";

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("tests the disabled flag and tooltip for each release status condition", async () => {
  type TProps = {
    code: hapi.release.Status.Code | null | undefined;
    disabled: boolean;
    tooltip?: string;
  };

  // this should cover all conditions
  const testsProps: TProps[] = [
    { code: 0, disabled: false, tooltip: undefined },
    { code: 1, disabled: false, tooltip: undefined },
    { code: 2, disabled: false, tooltip: undefined },
    { code: 3, disabled: false, tooltip: undefined },
    { code: 4, disabled: false, tooltip: undefined },
    { code: 5, disabled: true, tooltip: "deleted" },
    { code: 6, disabled: true, tooltip: "installation" },
    { code: 7, disabled: true, tooltip: "upgrade" },
    { code: 8, disabled: true, tooltip: "rollback" },
    { code: undefined, disabled: false, tooltip: undefined },
    { code: null, disabled: false, tooltip: undefined },
  ];

  for (const testProps of testsProps) {
    let releaseStatus;
    switch (testProps.code) {
      case null:
        releaseStatus = null;
        break;
      case undefined:
        releaseStatus = undefined;
        break;
      default:
        releaseStatus = {
          code: testProps.code,
        };
    }
    const disabled = testProps.disabled;
    const tooltip = testProps.tooltip ? testProps.tooltip : "";
    const wrapper = mountWrapper(
      defaultStore,
      <StatusAwareButton id="test" releaseStatus={releaseStatus} />,
    );

    // test disabled flag
    expect(wrapper.find(CdsButton).prop("disabled")).toBe(disabled);

    // test tooltip
    const tooltipUI = wrapper.find(ReactTooltip);
    if (disabled) {
      expect(tooltipUI).toExist();
      expect(tooltipUI).toIncludeText(tooltip);
    } else {
      expect(tooltipUI.exists()).toBeFalsy();
    }
  }
});
