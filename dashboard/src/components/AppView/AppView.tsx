import * as yaml from "js-yaml";
import * as React from "react";

import { Auth } from "../../shared/Auth";
import { ForbiddenError, IApp, IRBACRole, IResource, NotFoundError } from "../../shared/types";
import DeploymentStatus from "../DeploymentStatus";
import { NotFoundErrorAlert, PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";
import AppControls from "./AppControls";
import AppDetails from "./AppDetails";
import AppNotes from "./AppNotes";
import "./AppView.css";
import ChartInfo from "./ChartInfo";
import ServiceTable from "./ServiceTable";

interface IAppViewProps {
  namespace: string;
  releaseName: string;
  app: IApp;
  error: Error;
  deleteError: Error;
  getApp: (releaseName: string, namespace: string) => Promise<void>;
  deleteApp: (releaseName: string, namespace: string) => Promise<boolean>;
}

interface IAppViewState {
  deployments: Map<string, IResource>;
  otherResources: Map<string, IResource>;
  services: Map<string, IResource>;
  sockets: WebSocket[];
}

const RequiredRBACRoles: { [s: string]: IRBACRole[] } = {
  delete: [
    {
      apiGroup: "helm.bitnami.com",
      resource: "helmreleases",
      verbs: ["delete"],
    },
  ],
  view: [
    {
      apiGroup: "helm.bitnami.com",
      resource: "helmreleases",
      verbs: ["get"],
    },
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
    {
      apiGroup: "",
      namespace: "kubeapps",
      resource: "configmaps",
      verbs: ["get"],
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
    const manifest: IResource[] = yaml.safeLoadAll(newApp.data.manifest);
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
    const apiBase = `ws://${window.location.host}/api/kube`;
    const sockets: WebSocket[] = [];
    for (const d of deployments) {
      const s = new WebSocket(
        `${apiBase}/apis/apps/v1beta1/namespaces/${
          newApp.data.namespace
        }/deployments?watch=true&fieldSelector=metadata.name%3D${d.metadata.name}`,
        Auth.wsProtocols(),
      );
      s.addEventListener("message", e => this.handleEvent(e));
      sockets.push(s);
    }
    for (const svc of services) {
      const s = new WebSocket(
        `${apiBase}/api/v1/namespaces/${
          newApp.data.namespace
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
        this.setState({ services: { ...this.state.services, [key]: resource } });
        break;
    }
  }

  public render() {
    if (this.props.error) {
      return this.renderError(this.props.error);
    }
    if (!this.state.otherResources) {
      return <div>Loading</div>;
    }
    const { app } = this.props;
    if (!app) {
      return <div>Loading</div>;
    }
    return (
      <section className="AppView padding-b-big">
        <main>
          <div className="container">
            {this.props.deleteError && this.renderError(this.props.deleteError, "delete")}
            <div className="row collapse-b-tablet">
              <div className="col-3">
                <ChartInfo app={app} />
              </div>
              <div className="col-9">
                <div className="row padding-t-bigger">
                  <div className="col-4">
                    <DeploymentStatus deployments={this.deploymentArray()} />
                  </div>
                  <div className="col-8 text-r">
                    <AppControls app={app} deleteApp={this.deleteApp} />
                  </div>
                </div>
                <ServiceTable services={this.state.services} extended={false} />
                <AppNotes
                  notes={app.data.info && app.data.info.status && app.data.info.status.notes}
                />
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

  private renderError(error: Error, action: string = "view") {
    const { namespace, releaseName } = this.props;
    switch (error.constructor) {
      case ForbiddenError:
        return (
          <PermissionsErrorAlert
            namespace={namespace}
            roles={RequiredRBACRoles[action]}
            action={`${action} Application "${releaseName}"`}
          />
        );
      case NotFoundError:
        return (
          <NotFoundErrorAlert resource={`Application "${releaseName}"`} namespace={namespace} />
        );
      default:
        return <UnexpectedErrorAlert />;
    }
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

  private deleteApp = () => {
    return this.props.deleteApp(this.props.releaseName, this.props.namespace);
  };
}

export default AppView;
