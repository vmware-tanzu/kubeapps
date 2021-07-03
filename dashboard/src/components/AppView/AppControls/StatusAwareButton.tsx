import { CdsButton } from "@cds/react/button";
import ReactTooltip from "react-tooltip";
import { hapi } from "../../../shared/hapi/release";
import { inRange } from "lodash";

export interface IStatusAwareButtonProps {
  id: string;
  releaseStatus: hapi.release.IStatus | undefined | null;
}

export default function StatusAwareButton<T extends IStatusAwareButtonProps>(props: T) {
  const { id, releaseStatus, ...otherProps } = props;
  // Disable the button if: the status code is undefined or null OR the status code is one of [5,6,7,8]
  // See https://github.com/kubeapps/kubeapps/blob/master/dashboard/src/shared/hapi/release.d.ts#L559
  const disabled = releaseStatus?.code == null ? true : inRange(releaseStatus.code, 5, 9);
  const tooltips = {
    5: "The application is being deleted.",
    6: "The application is pending installation.",
    7: "The application is pending upgrade.",
    8: "The application is pending rollback.",
  };
  const tooltip = releaseStatus?.code ? tooltips[releaseStatus.code] : undefined;
  return (
    <>
      <CdsButton {...otherProps} disabled={disabled} data-for={id} data-tip={true} />
      {tooltip && (
        <ReactTooltip id={id} effect="solid" place="bottom">
          {tooltip}
        </ReactTooltip>
      )}
    </>
  );
}
