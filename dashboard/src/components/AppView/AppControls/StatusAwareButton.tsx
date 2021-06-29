import { CdsButton } from "@cds/react/button";
import ReactTooltip from "react-tooltip";
import { hapi } from "../../../shared/hapi/release";

export interface IStatusAwareButtonProps {
  id: string;
  releaseStatus: hapi.release.IStatus | undefined | null;
}

export default function StatusAwareButton<T extends IStatusAwareButtonProps>(props: T) {
  const { id, releaseStatus, ...otherProps } = props;
  const disabled: boolean = releaseStatus?.code
    ? releaseStatus.code >= 5 && releaseStatus.code <= 8
    : false;
  const tooltips = {
    5: "The applicatixon is being deleted.",
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
