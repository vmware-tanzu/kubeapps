import * as yaml from "js-yaml";
import * as React from "react";

import { IApp, IResource } from "../../shared/types";
import AppHeader from "./AppHeader";
import "./AppView.css";
import DeploymentWatcher from "./DeploymentWatcher";
import ServiceWatcher from "./ServiceWatcher";

interface IAppViewProps {
  namespace: string;
  releaseName: string;
  app: IApp;
  getApp: (releaseName: string) => Promise<void>;
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
    const { releaseName, getApp } = this.props;
    getApp(releaseName);
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
    const { releaseName } = this.props;

    if (!this.state.otherResources) {
      return <div>Loading</div>;
    }
    return (
      <section className="AppView padding-b-big">
        <AppHeader releasename={releaseName} />
        <main>
          <div className="container container-fluid">
            <DeploymentWatcher deployments={this.state.deployments} />
            <ServiceWatcher services={this.state.services} />
            <h6>Other Resources</h6>
            <table>
              <tbody>
                {this.state.otherResources &&
                  Object.keys(this.state.otherResources).map((k: string) => {
                    const r = this.state.otherResources[k];
                    return (
                      <tr key={k}>
                        <td>{r.kind}</td>
                        <td>{r.metadata.namespace}</td>
                        <td>{r.metadata.name}</td>
                      </tr>
                    );
                  })}
              </tbody>
            </table>
          </div>
        </main>
      </section>
    );
  }
}

export default AppView;
