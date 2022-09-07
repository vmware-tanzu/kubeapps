// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import { handleErrorAction } from "actions/auth";
import ErrorAlert from "components/ErrorAlert";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import PageHeader from "components/PageHeader/PageHeader";
import { push } from "connected-react-router";
import {
  InstalledPackageReference,
  ResourceRef,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import placeholder from "icons/placeholder.svg";
import * as yaml from "js-yaml";
import { useEffect, useMemo, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import * as ReactRouter from "react-router-dom";
import { Link } from "react-router-dom";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { InstalledPackage } from "shared/InstalledPackage";
import {
  CustomInstalledPackageDetail,
  DeleteError,
  FetchError,
  FetchWarning,
  IStoreState,
  NotFoundNetworkError,
} from "shared/types";
import { getPluginsSupportingRollback } from "shared/utils";
import ApplicationStatus from "../../containers/ApplicationStatusContainer";
import * as url from "../../shared/url";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import AccessURLTable from "./AccessURLTable/AccessURLTable";
import DeleteButton from "./AppControls/DeleteButton/DeleteButton";
import RollbackButton from "./AppControls/RollbackButton/RollbackButton";
import UpgradeButton from "./AppControls/UpgradeButton/UpgradeButton";
import AppNotes from "./AppNotes/AppNotes";
import AppSecrets from "./AppSecrets";
import AppValues from "./AppValues/AppValues";
import CustomAppView from "./CustomAppView";
import PackageInfo from "./PackageInfo/PackageInfo";
import ResourceTabs from "./ResourceTabs";

export interface IAppViewResourceRefs {
  deployments: ResourceRef[];
  statefulsets: ResourceRef[];
  daemonsets: ResourceRef[];
  services: ResourceRef[];
  ingresses: ResourceRef[];
  secrets: ResourceRef[];
  otherResources: ResourceRef[];
}

function parseResources(resourceRefs: Array<ResourceRef>) {
  const result: IAppViewResourceRefs = {
    ingresses: [],
    deployments: [],
    statefulsets: [],
    daemonsets: [],
    otherResources: [],
    services: [],
    secrets: [],
  };
  resourceRefs.forEach(ref => {
    switch (ref.kind) {
      case "Deployment":
        result.deployments.push(ref);
        break;
      case "StatefulSet":
        result.statefulsets.push(ref);
        break;
      case "DaemonSet":
        result.daemonsets.push(ref);
        break;
      case "Service":
        result.services.push(ref);
        break;
      case "Ingress":
        result.ingresses.push(ref);
        break;
      case "Secret":
        result.secrets.push(ref);
        break;
      default:
        result.otherResources.push(ref);
    }
  });
  return result;
}

function getButtons(installedPkg: CustomInstalledPackageDetail, error: any, revision: number) {
  if (
    !installedPkg ||
    !installedPkg?.installedPackageRef ||
    !installedPkg.installedPackageRef.plugin
  ) {
    return [];
  }

  const buttons = [];

  // Upgrade is a core operation, it will always be available
  buttons.push(
    <UpgradeButton
      key="upgrade-button"
      installedPackageRef={installedPkg.installedPackageRef}
      releaseStatus={installedPkg?.status}
      disabled={error !== undefined}
    />,
  );

  // Rollback is a helm-only operation, it will only be available for helm-plugin packages
  if (getPluginsSupportingRollback().includes(installedPkg.installedPackageRef.plugin.name)) {
    buttons.push(
      <RollbackButton
        key="rollback-button"
        installedPackageRef={installedPkg.installedPackageRef}
        revision={revision}
        releaseStatus={installedPkg?.status}
        disabled={error !== undefined}
      />,
    );
  }

  // Delete is a core operation, it will always be available
  buttons.push(
    <DeleteButton
      key="delete-button"
      installedPackageRef={installedPkg.installedPackageRef}
      releaseStatus={installedPkg?.status}
    />,
  );

  return buttons;
}

export interface IRouteParams {
  cluster: string;
  namespace: string;
  releaseName: string;
  pluginName: string;
  pluginVersion: string;
}

export default function AppView() {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const { cluster, namespace, releaseName, pluginName, pluginVersion } =
    ReactRouter.useParams() as IRouteParams;
  const [appViewResourceRefs, setAppViewResourceRefs] = useState({
    ingresses: [],
    deployments: [],
    statefulsets: [],
    daemonsets: [],
    otherResources: [],
    services: [],
    secrets: [],
  } as IAppViewResourceRefs);
  const {
    apps: {
      error,
      isFetching,
      selected: selectedInstalledPkg,
      selectedDetails: selectedAvailablePkg,
    },
    config: { customAppViews },
  } = useSelector((state: IStoreState) => state);

  const [fetchError, setFetchError] = useState(error);
  const [pluginObj] = useState({ name: pluginName, version: pluginVersion } as Plugin);
  const [resourceRefs, setResourceRefs] = useState([] as ResourceRef[]);

  // useMemo used so that when installedPkgRef is a dependency of other effects,
  // it does not trigger the effect on every render.
  const installedPkgRef = useMemo(() => {
    return {
      context: { cluster, namespace },
      identifier: releaseName,
      plugin: pluginObj,
    } as InstalledPackageReference;
  }, [cluster, namespace, releaseName, pluginObj]);

  useEffect(() => {
    dispatch(actions.installedpackages.getInstalledPackage(installedPkgRef));
  }, [dispatch, installedPkgRef]);

  useEffect(() => {
    // TODO(minelson): currently it is not possible for a client to determine
    // whether resource refs are unavailable because the package is being
    // installed (ie.  Package is "Pending" the actual installation) or the
    // package is currently unable to be installed because the RBAC isn't yet
    // correct (ie. Package is "Pending" required RBAC). The work-around here
    // is to continue polling for the resource refs every two seconds as long
    // as a `NotFound` is returned.
    // See https://github.com/vmware-tanzu/kubeapps/issues/4337
    let abort = false;
    const fetchResourceRefs = async () => {
      while (!abort) {
        try {
          const response = await InstalledPackage.GetInstalledPackageResourceRefs(installedPkgRef);
          if (abort) {
            return;
          }
          setResourceRefs(response.resourceRefs);
          return;
        } catch (e: any) {
          if (e.constructor !== NotFoundNetworkError) {
            // If we get any other error, we want the user to know about it.
            dispatch(handleErrorAction(e));
            setFetchError(new FetchError("unable to fetch resource references", [e]));
            return;
          }
          await new Promise(r => setTimeout(r, 2000));
        }
      }
    };
    fetchResourceRefs();

    // Ensure we abort fetching resource refs when unmounted.
    return () => {
      abort = true;
    };
  }, [dispatch, installedPkgRef]);

  useEffect(() => {
    if (resourceRefs.length === 0) {
      return () => {};
    }

    const parsedRefs = parseResources(resourceRefs);
    setAppViewResourceRefs(parsedRefs);
    return () => {};
  }, [resourceRefs]);

  useEffect(() => {
    if (!selectedInstalledPkg?.installedPackageRef) {
      return () => {};
    }
    const installedPackageRef = selectedInstalledPkg.installedPackageRef;
    // Watch Deployments, StatefulSets, DaemonSets, Ingresses and Services.
    const refsToWatch = appViewResourceRefs.deployments.concat(
      appViewResourceRefs.statefulsets,
      appViewResourceRefs.daemonsets,
      appViewResourceRefs.ingresses,
      appViewResourceRefs.services,
    );
    // And just get the rest.
    const refsToGet = appViewResourceRefs.secrets.concat(appViewResourceRefs.otherResources);
    if (refsToGet.length > 0) {
      dispatch(actions.kube.getResources(installedPackageRef, refsToGet, false));
    }
    if (refsToWatch.length > 0) {
      dispatch(actions.kube.getResources(installedPackageRef, refsToWatch, true));
      return function cleanup() {
        dispatch(actions.kube.closeRequestResources(installedPackageRef));
      };
    }
    return () => {};
  }, [dispatch, selectedInstalledPkg?.installedPackageRef, appViewResourceRefs]);

  const forceRetry = () => {
    dispatch(actions.installedpackages.clearErrorInstalledPackage());
    dispatch(actions.installedpackages.getInstalledPackage(installedPkgRef));
  };

  const goToAppsView = () => {
    dispatch(push(url.app.apps.list(cluster, namespace)));
  };

  if (fetchError) {
    if (fetchError.constructor === FetchError) {
      return (
        <ErrorAlert error={fetchError}>
          <CdsButton size="sm" action="flat" onClick={forceRetry} type="button">
            {" "}
            Try again{" "}
          </CdsButton>
        </ErrorAlert>
      );
    }
  }
  const { services, ingresses, deployments, statefulsets, daemonsets, secrets, otherResources } =
    appViewResourceRefs;
  const revision = selectedInstalledPkg?.revision ?? 0;
  const icon = selectedAvailablePkg?.iconUrl ?? placeholder;

  // If the package identifier matches the current list of loaded customAppViews,
  // then load the custom view from external bundle instead of the default one.
  const pkgRepo = selectedInstalledPkg?.availablePackageRef?.identifier.split("/")[0];
  const pkgName = selectedInstalledPkg?.availablePackageRef?.identifier.split("/")[1];
  const pkgPlugin = selectedInstalledPkg?.availablePackageRef?.plugin?.name;
  if (
    customAppViews.some(
      entry => entry.name === pkgName && entry.plugin === pkgPlugin && entry.repository === pkgRepo,
    )
  ) {
    return (
      <CustomAppView
        resourceRefs={appViewResourceRefs}
        app={selectedInstalledPkg!}
        appDetails={selectedAvailablePkg!}
      />
    );
  }
  return (
    <LoadingWrapper
      loaded={!isFetching}
      loadingText="Retrieving application..."
      className="margin-t-xl"
    >
      {!selectedInstalledPkg || !selectedInstalledPkg?.installedPackageRef ? (
        error ? (
          <Alert theme="danger">
            An error occurred while fetching the application: {error?.message}.{" "}
            <CdsButton size="sm" action="flat" onClick={goToAppsView} type="button">
              Go Back{" "}
            </CdsButton>
          </Alert>
        ) : (
          <></>
        )
      ) : (
        <section>
          <PageHeader
            title={releaseName}
            titleSize="md"
            subtitle={
              selectedAvailablePkg?.availablePackageRef ? (
                <span>
                  from package{" "}
                  <Link
                    to={url.app.packages.get(
                      cluster,
                      namespace,
                      selectedAvailablePkg.availablePackageRef,
                    )}
                  >
                    {selectedAvailablePkg.displayName}
                  </Link>
                </span>
              ) : (
                <span>from an unknown package</span>
              )
            }
            plugin={selectedInstalledPkg?.availablePackageRef?.plugin}
            icon={icon}
            buttons={getButtons(selectedInstalledPkg, error, revision)}
          />
          {error &&
            (error.constructor === FetchWarning ? (
              <Alert theme="warning">
                There is a problem with this package: {error["message"]}
              </Alert>
            ) : error.constructor === DeleteError ? (
              <Alert theme="danger">
                Unable to delete the application. Received: {error["message"]}
              </Alert>
            ) : (
              <Alert theme="danger">An error occurred: {error["message"]}</Alert>
            ))}
          {!selectedInstalledPkg || !selectedInstalledPkg?.status ? (
            <LoadingWrapper loadingText={`Loading ${releaseName}...`} />
          ) : (
            <Row>
              <Column span={3}>
                <PackageInfo
                  installedPackageDetail={selectedInstalledPkg}
                  availablePackageDetail={selectedAvailablePkg!}
                />
              </Column>
              <Column span={9}>
                <div className="appview-separator">
                  <div className="appview-first-row">
                    <ApplicationStatus
                      deployRefs={deployments}
                      statefulsetRefs={statefulsets}
                      daemonsetRefs={daemonsets}
                      info={selectedInstalledPkg}
                    />
                    <AccessURLTable serviceRefs={services} ingressRefs={ingresses} />
                    <AppSecrets secretRefs={secrets} />
                  </div>
                </div>
                <div className="appview-separator">
                  <AppNotes notes={selectedInstalledPkg?.postInstallationNotes} />
                </div>
                <div className="appview-separator">
                  <ResourceTabs
                    {...{
                      deployments,
                      statefulsets,
                      daemonsets,
                      secrets,
                      services,
                      otherResources,
                    }}
                  />
                </div>
                <div className="appview-separator">
                  <AppValues
                    values={
                      selectedInstalledPkg?.valuesApplied
                        ? yaml.dump(yaml.load(selectedInstalledPkg.valuesApplied))
                        : ""
                    }
                  />
                </div>
              </Column>
            </Row>
          )}
        </section>
      )}
    </LoadingWrapper>
  );
}
