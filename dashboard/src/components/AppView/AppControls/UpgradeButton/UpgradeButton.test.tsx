import { CdsButton } from "@cds/react/button";
import actions from "actions";
import {
  InstalledPackageStatus_StatusReason,
  InstalledPackageStatus,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import * as ReactRedux from "react-redux";
import ReactTooltip from "react-tooltip";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import UpgradeButton from "./UpgradeButton";

const defaultProps = {
  cluster: "default",
  namespace: "kubeapps",
  releaseName: "foo",
  releaseStatus: null,
};

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.apps = {
    ...actions.apps,
    upgradeApp: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("should render a disabled button if when passing an in-progress status", async () => {
  const disabledProps = {
    ...defaultProps,
    releaseStatus: {
      ready: false,
      reason: InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING,
      userReason: "Pending",
    } as InstalledPackageStatus,
  };
  const wrapper = mountWrapper(defaultStore, <UpgradeButton {...disabledProps} />);

  expect(wrapper.find(CdsButton)).toBeDisabled();
  expect(wrapper.find(ReactTooltip)).toExist();
});
