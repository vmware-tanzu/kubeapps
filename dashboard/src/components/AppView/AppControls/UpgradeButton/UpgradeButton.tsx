import { Link } from "react-router-dom";
import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import ReactTooltip from "react-tooltip";
import * as url from "../../../../shared/url";
import { hapi } from "../../../../shared/hapi/release";

interface IUpgradeButtonProps {
  cluster: string;
  namespace: string;
  releaseName: string;
  status: hapi.release.IStatus | undefined | null;
}

export default function UpgradeButton({
  cluster,
  namespace,
  releaseName,
  status,
}: IUpgradeButtonProps) {
  const disabled: boolean = status?.code ? status.code >= 5 && status.code <= 8 : false;
  const tooltips = {
    5: "The application is being deleted.",
    6: "The application is pending installation.",
    7: "The application is pending upgrade.",
    8: "The application is pending rollback.",
  };
  const tooltip = status?.code ? tooltips[status.code] : undefined;
  return (
    <Link to={url.app.apps.upgrade(cluster, namespace, releaseName)}>
      <CdsButton status="primary" disabled={disabled} data-tip={true} data-for="upgrade-button">
        <CdsIcon shape="upload-cloud" /> Upgrade
      </CdsButton>
      {tooltip && (
        <ReactTooltip id="upgrade-button" effect="solid" place="bottom">
          {tooltip}
        </ReactTooltip>
      )}
    </Link>
  );
}
