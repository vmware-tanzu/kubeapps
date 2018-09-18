import * as React from "react";
import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { withRouter } from "react-router";

import Routes from "../../components/Routes";
import { IStoreState } from "../../shared/types";
import AppList from "../AppListContainer";
import AppNew from "../AppNewContainer";
import AppUpgrade from "../AppUpgradeContainer";
import AppView from "../AppViewContainer";
import ChartList from "../ChartListContainer";
import ChartView from "../ChartViewContainer";
import ClassListContainer from "../ClassListContainer";
import { ClassViewContainer } from "../ClassView";
import FunctionListContainer from "../FunctionListContainer";
import FunctionViewContainer from "../FunctionViewContainer";
import InstanceListViewContainer from "../InstanceListViewContainer";
import InstanceView from "../InstanceView";
import LoginFormContainer from "../LoginFormContainer";
import RepoListContainer from "../RepoListContainer";
import ServiceCatalogContainer from "../ServiceCatalogContainer";

const privateRoutes: {
  [route: string]: React.ComponentType<RouteComponentProps<any>> | React.ComponentType<any>;
} = {
  "/apps/ns/:namespace": AppList,
  "/apps/ns/:namespace/:releaseName": AppView,
  "/apps/ns/:namespace/new/:repo/:id/versions/:version": AppNew,
  "/apps/ns/:namespace/upgrade/:releaseName": AppUpgrade,
  "/charts": ChartList,
  "/charts/:repo": ChartList,
  "/charts/:repo/:id": ChartView,
  "/charts/:repo/:id/versions/:version": ChartView,
  "/config/brokers": ServiceCatalogContainer,
  "/config/repos": RepoListContainer,
  "/functions/ns/:namespace": FunctionListContainer,
  "/functions/ns/:namespace/:name": FunctionViewContainer,
  "/services/brokers/:brokerName/classes/:className": ClassViewContainer,
  "/services/brokers/:brokerName/instances/ns/:namespace/:instanceName": InstanceView,
  "/services/classes": ClassListContainer,
  "/services/instances/ns/:namespace": InstanceListViewContainer,
};

// Public routes that don't require authentication
const routes: {
  [route: string]: React.ComponentType<RouteComponentProps<any>> | React.ComponentType<any>;
} = {
  "/login": LoginFormContainer,
};

function mapStateToProps({ namespace }: IStoreState) {
  return { namespace: namespace.current, privateRoutes, routes };
}

export default withRouter(connect(mapStateToProps)(Routes));
