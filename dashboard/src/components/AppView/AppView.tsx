import * as yaml from "js-yaml";
import * as React from "react";

import { IApp, IResource } from "../../shared/types";
import AppControls from "./AppControls";
import AppDetails from "./AppDetails";
import AppNotes from "./AppNotes";
import AppStatus from "./AppStatus";
import "./AppView.css";
import ChartInfo from "./ChartInfo";
import ServiceTable from "./ServiceTable";

interface IAppViewProps {
  namespace: string;
  releaseName: string;
  app: IApp;
  getApp: (releaseName: string, namespace: string) => Promise<void>;
  deleteApp: (releaseName: string, namespace: string) => Promise<void>;
}

interface IAppViewState {
  deployments: Map<string, IResource>;
  otherResources: Map<string, IResource>;
  services: Map<string, IResource>;
  sockets: WebSocket[];
}

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

  public componentWillReceiveProps(nextProps: IAppViewProps) {
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
      );
      s.addEventListener("message", e => this.handleEvent(e));
      sockets.push(s);
    }
    for (const svc of services) {
      const s = new WebSocket(
        `${apiBase}/api/v1/namespaces/${
          newApp.data.namespace
        }/services?watch=true&fieldSelector=metadata.name%3D${svc.metadata.name}`,
      );
      s.addEventListener("message", e => this.handleEvent(e));
      sockets.push(s);
    }
    this.setState({
      sockets,
    });
  }

  public componentWillUnmount() {
    const { sockets } = this.state;
    for (const s of sockets) {
      s.close();
    }
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
            <div className="row collapse-b-tablet">
              <div className="col-3">
                <ChartInfo app={app} />
              </div>
              <div className="col-9">
                <div className="row padding-t-bigger">
                  <div className="col-4">
                    <AppStatus deployments={this.state.deployments} />
                  </div>
                  <div className="col-4" />
                  <div className="col-4">
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

  public deleteApp = () => {
    return this.props.deleteApp(this.props.releaseName, this.props.namespace);
  };
}

export default AppView;
