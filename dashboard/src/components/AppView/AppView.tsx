import * as yaml from "js-yaml";
import * as _ from "lodash";
import * as React from "react";

import SecretTable from "../../containers/SecretsTableContainer";
import { Auth } from "../../shared/Auth";
import { hapi } from "../../shared/hapi/release";
import { IK8sList, IKubeItem, IRBACRole, IResource } from "../../shared/types";
import WebSocketHelper from "../../shared/WebSocketHelper";
import DeploymentStatus from "../DeploymentStatus";
import { ErrorSelector } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import AccessURLTable from "./AccessURLTable";
import AppControls from "./AppControls";
import AppNotes from "./AppNotes";
import "./AppView.css";
import ChartInfo from "./ChartInfo";
import DeploymentsTable from "./DeploymentsTable";
import OtherResourcesTable from "./OtherResourcesTable";
import ServicesTable from "./ServicesTable";

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
  deployments: Array<IKubeItem<IResource>>;
  services: Array<IKubeItem<IResource>>;
  ingresses: Array<IKubeItem<IResource>>;
  // Other resources are not IKubeItems because
  // we are not fetching any information for them.
  otherResources: IResource[];
  secretNames: string[];
  sockets: WebSocket[];
  manifest: IResource[];
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
    manifest: [],
    deployments: [],
    ingresses: [],
    otherResources: [],
    services: [],
    secretNames: [],
    sockets: [],
  };

  public async componentDidMount() {
    const { releaseName, getApp, namespace } = this.props;
    getApp(releaseName, namespace);
  }

  public componentWillReceiveProps(nextProps: IAppViewProps) {
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
    // TODO(prydonius): Okay to use non-safe load here since we assume the
    // manifest is pre-parsed by Helm and Kubernetes. Look into switching back
    // to safeLoadAll once https://github.com/nodeca/js-yaml/issues/456 is
    // resolved.
    let manifest: IResource[] = yaml.loadAll(newApp.manifest, undefined, { json: true });
    // Filter out elements in the manifest that does not comply
    // with { kind: foo }
    manifest = manifest.filter(r => r && r.kind);
    if (!_.isEqual(manifest, this.state.manifest)) {
      this.setState({ manifest });
    } else {
      return;
    }

    const sockets: WebSocket[] = [];
    // Iterate over the current manifest to populate the initial state
    const secretNames: string[] = [];
    const deployments: Array<IKubeItem<IResource>> = [];
    const services: Array<IKubeItem<IResource>> = [];
    const ingresses: Array<IKubeItem<IResource>> = [];
    const otherResources: IResource[] = [];
    manifest.forEach((i: IResource | IK8sList<IResource, {}>) => {
      const item = i as IResource;
      const resource = { isFetching: true, item };
      switch (i.kind) {
        case "Deployment":
          deployments.push(resource);
          sockets.push(
            this.getSocket("deployments", i.apiVersion, item.metadata.name, newApp.namespace),
          );
          break;
        case "Service":
          services.push(resource);
          sockets.push(
            this.getSocket("services", i.apiVersion, item.metadata.name, newApp.namespace),
          );
          break;
        case "Ingress":
          ingresses.push(resource);
          sockets.push(
            this.getSocket("ingresses", i.apiVersion, item.metadata.name, newApp.namespace),
          );
          break;
        case "Secret":
          secretNames.push(item.metadata.name);
          break;
        case "List":
          (i as IK8sList<IResource, {}>).items.forEach(listItem => otherResources.push(listItem));
          break;
        default:
          otherResources.push(item);
      }
    });
    this.setState({
      sockets,
      deployments,
      services,
      ingresses,
      secretNames,
      otherResources,
    });
  }

  public componentWillUnmount() {
    this.closeSockets();
  }

  public handleEvent(e: MessageEvent) {
    const msg = JSON.parse(e.data);
    const resource: IResource = msg.object;
    const newItem = {
      isFetching: false,
      item: resource,
    };
    const dropByName = (array: Array<IKubeItem<IResource>>) => {
      return _.dropWhile(array, r => r.item && r.item.metadata.name === resource.metadata.name);
    };
    switch (resource.kind) {
      case "Deployment":
        const newDeps = dropByName(this.state.deployments);
        newDeps.push(newItem);
        this.setState({ deployments: newDeps });
        break;
      case "Service":
        const newSvcs = dropByName(this.state.services);
        newSvcs.push(newItem);
        this.setState({ services: newSvcs });
        break;
      case "Ingress":
        const newIngresses = dropByName(this.state.ingresses);
        newIngresses.push(newItem);
        this.setState({ ingresses: newIngresses });
        break;
    }
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

    return this.props.app && this.props.app.info ? this.appInfo() : <LoadingWrapper />;
  }

  public appInfo() {
    const { app } = this.props;
    const { services, ingresses, deployments, secretNames, otherResources } = this.state;
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
                    <DeploymentStatus deployments={deployments} info={app.info!} />
                  </div>
                  <div className="col-8 text-r">
                    <AppControls app={app} deleteApp={this.deleteApp} />
                  </div>
                </div>
                <AccessURLTable services={services} ingresses={ingresses} />
                <AppNotes notes={app.info && app.info.status && app.info.status.notes} />
                <SecretTable namespace={app.namespace} secretNames={secretNames} />
                <DeploymentsTable deployments={deployments} />
                <ServicesTable services={services} />
                <OtherResourcesTable otherResources={otherResources} />
              </div>
            </div>
          </div>
        </main>
      </section>
    );
  }

  private getSocket(
    resource: string,
    apiVersion: string,
    name: string,
    namespace: string,
  ): WebSocket {
    const apiBase = WebSocketHelper.apiBase();
    const s = new WebSocket(
      `${apiBase}/${
        apiVersion === "v1" ? "api/v1" : `apis/${apiVersion}`
      }/namespaces/${namespace}/${resource}?watch=true&fieldSelector=metadata.name%3D${name}`,
      Auth.wsProtocols(),
    );
    s.addEventListener("message", e => this.handleEvent(e));
    return s;
  }

  private closeSockets() {
    const { sockets } = this.state;
    for (const s of sockets) {
      s.close();
    }
  }

  private deleteApp = (purge: boolean) => {
    return this.props.deleteApp(this.props.releaseName, this.props.namespace, purge);
  };
}

export default AppView;
