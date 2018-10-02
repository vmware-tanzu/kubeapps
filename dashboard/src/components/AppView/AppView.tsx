import * as yaml from "js-yaml";
import * as React from "react";

import { Auth } from "../../shared/Auth";
import { hapi } from "../../shared/hapi/release";
import { IRBACRole, IResource } from "../../shared/types";
import WebSocketHelper from "../../shared/WebSocketHelper";
import DeploymentStatus from "../DeploymentStatus";
import { ErrorSelector } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import AccessURLTable from "./AccessURLTable";
import AppControls from "./AppControls";
import AppDetails from "./AppDetails";
import AppNotes from "./AppNotes";
import "./AppView.css";
import ChartInfo from "./ChartInfo";

export interface IAppViewProps {
  namespace: string;
  releaseName: string;
  app: hapi.release.Release;
  // TODO(miguel) how to make optional props? I tried adding error? but the container complains
  error: Error | undefined;
  deleteError: Error | undefined;
  getApp: (releaseName: string, namespace: string) => void;
  deleteApp: (releaseName: string, namespace: string, purge: boolean) => Promise<boolean>;
}

interface IAppViewState {
  deployments: Map<string, IResource>;
  otherResources: Map<string, IResource>;
  services: Map<string, IResource>;
  sockets: WebSocket[];
}

const RequiredRBACRoles: { [s: string]: IRBACRole[] } = {
  view: [
    {
      apiGroup: "apps",
      resource: "deployments",
      verbs: ["list", "watch"],
    },
    {
      apiGroup: "apps",
      resource: "services",
      verbs: ["list", "watch"],
    },
  ],
};

class AppView extends React.Component<IAppViewProps, IAppViewState> {
  public state: IAppViewState = {
    deployments: new Map<string, IResource>(),
    otherResources: new Map<string, IResource>(),
    services: new Map<string, IResource>(),
    sockets: [],
  };

  public async componentDidMount() {
    const { releaseName, getApp, namespace } = this.props;
    getApp(releaseName, namespace);
  }

  public async componentWillReceiveProps(nextProps: IAppViewProps) {
    const { releaseName, getApp, namespace } = this.props;
    if (nextProps.namespace !== namespace) {
      getApp(releaseName, nextProps.namespace);
      return;
    }
    if (nextProps.error) {
      // close any existing sockets
      this.closeSockets();
      return;
    }
    const newApp = nextProps.app;
    if (!newApp) {
      return;
    }
    let manifest: IResource[] = yaml.safeLoadAll(newApp.manifest);
    // Filter out elements in the manifest that does not comply
    // with { kind: foo }
    manifest = manifest.filter(r => r && r.kind);

    const watchedKinds = ["Deployment", "Service"];
    const otherResources = manifest
      .filter(d => watchedKinds.indexOf(d.kind) < 0)
      .reduce((acc, r) => {
        // TODO: skip list resource for now
        if (r.kind === "List") {
          return acc;
        }
        acc[`${r.kind}/${r.metadata.name}`] = r;
        return acc;
      }, new Map<string, IResource>());
    this.setState({ otherResources });

    const deployments = manifest.filter(d => d.kind === "Deployment");
    const services = manifest.filter(d => d.kind === "Service");
    const apiBase = WebSocketHelper.apiBase();
    const sockets: WebSocket[] = [];
    for (const d of deployments) {
      const s = new WebSocket(
        `${apiBase}/apis/apps/v1beta1/namespaces/${
          newApp.namespace
        }/deployments?watch=true&fieldSelector=metadata.name%3D${d.metadata.name}`,
        Auth.wsProtocols(),
      );
      s.addEventListener("message", e => this.handleEvent(e));
      sockets.push(s);
    }
    for (const svc of services) {
      const s = new WebSocket(
        `${apiBase}/api/v1/namespaces/${
          newApp.namespace
        }/services?watch=true&fieldSelector=metadata.name%3D${svc.metadata.name}`,
        Auth.wsProtocols(),
      );
      s.addEventListener("message", e => this.handleEvent(e));
      sockets.push(s);
    }
    this.setState({
      sockets,
    });
  }

  public componentWillUnmount() {
    this.closeSockets();
  }

  public handleEvent(e: MessageEvent) {
    const msg = JSON.parse(e.data);
    const resource: IResource = msg.object;
    const key = `${resource.kind}/${resource.metadata.name}`;
    switch (resource.kind) {
      case "Deployment":
        this.setState({ deployments: { ...this.state.deployments, [key]: resource } });
        break;
      case "Service":
        this.setState({ services: this.state.services.set(key, resource) });
        break;
    }
  }

  public get isLoading(): boolean {
    const { app } = this.props;
    return !this.state.otherResources || (!app || !app.info);
  }

  public render() {
    if (this.props.error) {
      return (
        <ErrorSelector
          error={this.props.error}
          defaultRequiredRBACRoles={RequiredRBACRoles}
          action="view"
          resource={`Application ${this.props.releaseName}`}
          namespace={this.props.namespace}
        />
      );
    }

    return this.isLoading ? <LoadingWrapper /> : this.appInfo();
  }

  public appInfo() {
    const { app } = this.props;
    // Although LoadingWrapper checks that the app is loaded before loading this wrapper
    // it seems that react renders it even before causing it to crash because app is null
    // that's why we need to have an app && guard clause
    return (
      <section className="AppView padding-b-big">
        <main>
          <div className="container">
            {this.props.deleteError && (
              <ErrorSelector
                error={this.props.deleteError}
                defaultRequiredRBACRoles={RequiredRBACRoles}
                action="delete"
                resource={`Application ${this.props.releaseName}`}
                namespace={this.props.namespace}
              />
            )}
            <div className="row collapse-b-tablet">
              <div className="col-3">
                <ChartInfo app={app} />
              </div>
              <div className="col-9">
                <div className="row padding-t-bigger">
                  <div className="col-4">
                    <DeploymentStatus deployments={this.deploymentArray()} info={app.info!} />
                  </div>
                  <div className="col-8 text-r">
                    <AppControls app={app} deleteApp={this.deleteApp} />
                  </div>
                </div>
                {this.state.services.size > 0 && <AccessURLTable services={this.state.services} />}
                <AppNotes notes={app.info && app.info.status && app.info.status.notes} />
                <AppDetails
                  deployments={this.state.deployments}
                  services={this.state.services}
                  otherResources={this.state.otherResources}
                />
              </div>
            </div>
          </div>
        </main>
      </section>
    );
  }

  private closeSockets() {
    const { sockets } = this.state;
    for (const s of sockets) {
      s.close();
    }
  }

  private deploymentArray(): IResource[] {
    return Object.keys(this.state.deployments).map(k => this.state.deployments[k]);
  }

  private deleteApp = (purge: boolean) => {
    return this.props.deleteApp(this.props.releaseName, this.props.namespace, purge);
  };
}

export default AppView;
