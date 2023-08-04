// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import AppList from "components/AppList/AppList";
import AppUpgrade from "components/AppUpgrade";
import AppView from "components/AppView";
import Catalog from "components/Catalog/Catalog";
import DeploymentForm from "components/DeploymentForm";
import LoadingWrapper from "components/LoadingWrapper";
import PackageView from "components/PackageHeader";
import { Navigate, Route, Routes } from "react-router-dom";
import { app } from "shared/url";
import ApiDocs from "../../components/ApiDocs";
import NotFound from "../../components/NotFound";
import AlertGroup from "components/AlertGroup";
import PkgRepoList from "components/Config/PkgRepoList/PkgRepoList";

import OperatorNew from "components/OperatorNew";
import OperatorInstanceForm from "components/OperatorInstanceForm";
import OperatorList from "components/OperatorList";
import OperatorView from "components/OperatorView";
import OperatorInstance from "components/OperatorInstance";
import OperatorInstanceUpdateForm from "components/OperatorInstanceUpdateForm";
import LoginForm from "components/LoginForm";
import RequireAuthentication from "components/RequireAuthentication";
import { useSelector } from "react-redux";
import { IStoreState } from "shared/types";

const privateRoutes = {
  "/c/:cluster/ns/:namespace/apps": AppList,
  "/c/:cluster/ns/:namespace/apps/:pluginName/:pluginVersion/:releaseName": AppView,
  "/c/:cluster/ns/:namespace/apps/:pluginName/:pluginVersion/:releaseName/upgrade": AppUpgrade,
  "/c/:cluster/ns/:namespace/apps/:pluginName/:pluginVersion/:releaseName/upgrade/:version":
    AppUpgrade,
  "/c/:cluster/ns/:namespace/apps/new/:pluginName/:pluginVersion/:packageCluster/:packageNamespace/:packageId/versions/:packageVersion":
    DeploymentForm,
  AppUpgrade,
  "/c/:cluster/ns/:namespace/apps/new/:pluginName/:pluginVersion/:packageCluster/:packageNamespace/:packageId/versions":
    DeploymentForm,
  "/c/:cluster/ns/:namespace/catalog": Catalog,
  "/c/:cluster/ns/:namespace/packages/:pluginName/:pluginVersion/:packageCluster/:packageNamespace/:packageId":
    PackageView,
  "/c/:cluster/ns/:namespace/packages/:pluginName/:pluginVersion/:packageCluster/:packageNamespace/:packageId/versions/:packageVersion":
    PackageView,
  "/c/:cluster/ns/:namespace/config/repos": PkgRepoList,
  "/docs": ApiDocs,
} as const;

const operatorsRoutes = {
  "/c/:cluster/ns/:namespace/operators": OperatorList,
  "/c/:cluster/ns/:namespace/operators/:operator": OperatorView,
  "/c/:cluster/ns/:namespace/operators/new/:operator": OperatorNew,
  "/c/:cluster/ns/:namespace/operators-instances/new/:csv/:crd": OperatorInstanceForm,
  "/c/:cluster/ns/:namespace/operators-instances/:csv/:crd/:instanceName": OperatorInstance,
  "/c/:cluster/ns/:namespace/operators-instances/:csv/:crd/:instanceName/update":
    OperatorInstanceUpdateForm,
} as const;

const unsupportedRoutes = {
  "/c/:cluster/ns/:namespace/operators/*":
    "Operators support has been deactivated by default for Kubeapps. It can be enabled in values configuration.",
  "/c/:cluster/ns/:namespace/operators-instances/*":
    "Operators support has been deactivated by default for Kubeapps. It can be enabled in values configuration.",
} as const;

function AppRoutes() {
  const {
    config: { featureFlags },
    clusters: { currentCluster: cluster, clusters },
    auth: { authenticated },
  } = useSelector((state: IStoreState) => state);
  const currentNamespace = clusters[cluster].currentNamespace;
  const rootNamespacedRedirect = () => {
    if (authenticated) {
      if (!cluster || !currentNamespace) {
        return <LoadingWrapper className="margin-t-xxl" loadingText="Fetching Cluster Info..." />;
      }
      return <Navigate replace to={{ pathname: app.apps.list(cluster, currentNamespace) }} />;
    }
    // There is not a default namespace, redirect to login page
    return <Navigate replace to={{ pathname: "/login" }} />;
  };
  return (
    <Routes>
      <Route path="/" element={rootNamespacedRedirect()} />
      <Route key="/login" path="/login" element={<LoginForm />} />
      {Object.entries(privateRoutes).map(([route, component]) => {
        const Component = component;
        return (
          <Route
            key={route}
            path={route}
            element={
              <RequireAuthentication>
                <Component />
              </RequireAuthentication>
            }
          />
        );
      })}
      {featureFlags?.operators &&
        Object.entries(operatorsRoutes).map(([route, component]) => {
          const Component = component;
          return (
            <Route
              key={route}
              path={route}
              element={
                <RequireAuthentication>
                  <Component />
                </RequireAuthentication>
              }
            />
          );
        })}
      {!featureFlags?.operators &&
        Object.entries(unsupportedRoutes).map(([route, message]) => {
          return (
            <Route
              key={route}
              path={route}
              element={
                <div className="margin-t-sm">
                  <AlertGroup status="warning">{message}</AlertGroup>
                </div>
              }
            />
          );
        })}
      {/* If the route doesn't match any expected path redirect to a 404 page  */}
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
}

export default AppRoutes;
