import { CdsButton } from "@cds/react/button";
import { hapi } from "../../../shared/hapi/release";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import StatusAwareButton from "./StatusAwareButton";
import ReactTooltip from "react-tooltip";

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
    { code: undefined, disabled: true, tooltip: undefined },
    { code: null, disabled: true, tooltip: undefined },
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
    const tooltip = testProps.tooltip;
    const wrapper = mountWrapper(
      defaultStore,
      <StatusAwareButton id="test" releaseStatus={releaseStatus} />,
    );

    // test disabled flag
    expect(wrapper.find(CdsButton).prop("disabled")).toBe(disabled);

    // test tooltip
    const tooltipUI = wrapper.find(ReactTooltip);
    if (tooltip) {
      expect(tooltipUI).toExist();
      expect(tooltipUI).toIncludeText(tooltip);
    } else {
      expect(tooltipUI.exists()).toBeFalsy();
    }
  }
});
