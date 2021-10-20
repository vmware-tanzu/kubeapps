import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import AvailablePackageDetailExcerpt from "components/Catalog/AvailablePackageDetailExcerpt";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import LoadingWrapper from "components/LoadingWrapper";
import { push } from "connected-react-router";
import { AvailablePackageReference } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import * as ReactRouter from "react-router";
import { Link } from "react-router-dom";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import { app } from "shared/url";
import PackageHeader from "./PackageHeader";
import PackageReadme from "./PackageReadme";

interface IRouteParams {
  cluster: string;
  namespace: string;
  global: string;
  pluginName: string;
  pluginVersion: string;
  packageId: string;
  packageVersion?: string;
}

export default function PackageView() {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const location = ReactRouter.useLocation();
  const {
    cluster: targetCluster,
    namespace: targetNamespace,
    global,
    packageId,
    pluginName,
    pluginVersion,
    packageVersion,
  } = ReactRouter.useParams() as IRouteParams;
  const {
    config,
    packages: { isFetching, selected: selectedPackage },
  } = useSelector((state: IStoreState) => state);

  const [pluginObj] = useState({ name: pluginName, version: pluginVersion } as Plugin);

  const isGlobal = global === "global";

  // Use the cluster/namespace from the URL unless it comes from a "global" repository.
  // In that case, use the cluster/namespace from where kubeapps has been installed on
  const packageCluster = isGlobal ? config.kubeappsCluster : targetCluster;
  const packageNamespace = isGlobal ? config.kubeappsNamespace : targetNamespace;

  // Fetch the selected/latest version on the initial load
  useEffect(() => {
    dispatch(
      actions.packages.fetchAndSelectAvailablePackageDetail(
        {
          context: { cluster: packageCluster, namespace: packageNamespace },
          plugin: pluginObj,
          identifier: packageId,
        } as AvailablePackageReference,
        packageVersion,
      ),
    );
    return () => {};
  }, [dispatch, packageId, packageNamespace, packageCluster, packageVersion, pluginObj]);

  // Fetch all versions
  useEffect(() => {
    dispatch(
      actions.packages.fetchAvailablePackageVersions({
        context: { cluster: packageCluster, namespace: packageNamespace },
        plugin: { name: pluginName, version: pluginVersion } as Plugin,
        identifier: packageId,
      } as AvailablePackageReference),
    );
    return () => {};
  }, [dispatch, packageId, packageNamespace, packageCluster, pluginName, pluginVersion]);

  // Select version handler
  const selectVersion = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const versionRegex = /\/versions\/(.*)/;
    if (versionRegex.test(location.pathname)) {
      // If the current URL already has the version, replace it
      dispatch(push(location.pathname.replace(versionRegex, `/versions/${event.target.value}`)));
    } else {
      // Otherwise, append the version
      const trimmedPath = location.pathname.endsWith("/")
        ? location.pathname.slice(0, -1)
        : location.pathname;
      dispatch(push(trimmedPath.concat(`/versions/${event.target.value}`)));
    }
  };

  if (selectedPackage.error) {
    return <Alert theme="danger">Unable to fetch package: {selectedPackage.error.message}</Alert>;
  }
  if (isFetching || !selectedPackage.availablePackageDetail || !selectedPackage.pkgVersion) {
    return <LoadingWrapper loaded={false} />;
  }
  return (
    <section>
      <div>
        <PackageHeader
          availablePackageDetail={selectedPackage.availablePackageDetail}
          versions={selectedPackage.versions}
          onSelect={selectVersion}
          deployButton={
            <Link
              to={app.apps.new(
                targetCluster,
                targetNamespace,
                pluginObj,
                packageId,
                selectedPackage.pkgVersion,
                isGlobal,
              )}
            >
              <CdsButton status="primary">
                <CdsIcon shape="deploy" /> Deploy
              </CdsButton>
            </Link>
          }
          selectedVersion={selectedPackage.pkgVersion}
        />
      </div>

      <section>
        <Row>
          <Column span={3}>
            <AvailablePackageDetailExcerpt pkg={selectedPackage.availablePackageDetail} />
          </Column>
          <Column span={9}>
            <PackageReadme readme={selectedPackage.readme} error={selectedPackage.readmeError} />
            <div className="after-readme-button">
              <Link
                to={app.apps.new(
                  targetCluster,
                  targetNamespace,
                  pluginObj,
                  packageId,
                  selectedPackage.pkgVersion,
                  isGlobal,
                )}
              >
                <CdsButton status="primary">
                  <CdsIcon shape="deploy" /> Deploy
                </CdsButton>
              </Link>
            </div>
          </Column>
        </Row>
      </section>
    </section>
  );
}
