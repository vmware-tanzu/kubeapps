import { CdsIcon } from "@cds/react/icon";
import { InstalledPackageStatus } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Link } from "react-router-dom";
import * as url from "../../../../shared/url";
import StatusAwareButton from "../StatusAwareButton";

interface IUpgradeButtonProps {
  cluster: string;
  namespace: string;
  releaseName: string;
  releaseStatus: InstalledPackageStatus | undefined | null;
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
