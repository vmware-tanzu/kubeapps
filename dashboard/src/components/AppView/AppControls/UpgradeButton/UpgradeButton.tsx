import { Link } from "react-router-dom";
import { CdsIcon } from "@cds/react/icon";
import StatusAwareButton from "../StatusAwareButton";
import * as url from "../../../../shared/url";
import { hapi } from "../../../../shared/hapi/release";

interface IUpgradeButtonProps {
  cluster: string;
  namespace: string;
  releaseName: string;
  releaseStatus: hapi.release.IStatus | undefined | null;
}

export default function UpgradeButton({
  cluster,
  namespace,
  releaseName,
  releaseStatus,
}: IUpgradeButtonProps) {
  return (
    <Link to={url.app.apps.upgrade(cluster, namespace, releaseName)}>
      <StatusAwareButton id="upgrade-button" status="primary" releaseStatus={releaseStatus}>
        <CdsIcon shape="upload-cloud" /> Upgrade
      </StatusAwareButton>
    </Link>
  );
}
