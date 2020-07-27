import { RouterAction } from "connected-react-router";
import { assignWith, isEqual } from "lodash";
import * as React from "react";
import * as yaml from "yaml";
import { hapi } from "../../shared/hapi/release";

import AccessURLTable from "../../containers/AccessURLTableContainer";
import ApplicationStatus from "../../containers/ApplicationStatusContainer";
import ResourceRef from "../../shared/ResourceRef";
import { IK8sList, IRBACRole, IRelease, IResource } from "../../shared/types";
import { ErrorSelector } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import AppControls from "./AppControls";
import AppNotes from "./AppNotes";
import AppValues from "./AppValues";
import "./AppView.css";
import ChartInfo from "./ChartInfo";
import ResourceTable from "./ResourceTable";

export interface IAppViewProps {
  cluster: string;
  namespace: string;
  releaseName: string;
  app?: IRelease;
  // TODO(miguel) how to make optional props? I tried adding error? but the container complains
  error: Error | undefined;
  deleteError: Error | undefined;
  getAppWithUpdateInfo: (cluster: string, namespace: string, releaseName: string) => void;
  deleteApp: (
    cluster: string,
    namespace: string,
    releaseName: string,
    purge: boolean,
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
}

interface IAppViewState {
  deployRefs: ResourceRef[];
  statefulSetRefs: ResourceRef[];
  daemonSetRefs: ResourceRef[];
  serviceRefs: ResourceRef[];
  ingressRefs: ResourceRef[];
  secretRefs: ResourceRef[];
  otherResources: ResourceRef[];
  manifest: IResource[];
}

export interface IPartialAppViewState {
  deployRefs: ResourceRef[];
  statefulSetRefs: ResourceRef[];
  daemonSetRefs: ResourceRef[];
  serviceRefs: ResourceRef[];
  ingressRefs: ResourceRef[];
  secretRefs: ResourceRef[];
  otherResources: ResourceRef[];
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
    ingressRefs: [],
    deployRefs: [],
    statefulSetRefs: [],
    daemonSetRefs: [],
    otherResources: [],
    serviceRefs: [],
    secretRefs: [],
  };

  public async componentDidMount() {
    const { releaseName, getAppWithUpdateInfo, cluster, namespace } = this.props;
    getAppWithUpdateInfo(cluster, namespace, releaseName);
  }

  public componentDidUpdate(prevProps: IAppViewProps) {
    const { releaseName, getAppWithUpdateInfo, cluster, namespace, error, app } = this.props;
    if (prevProps.namespace !== namespace || prevProps.cluster !== cluster) {
      getAppWithUpdateInfo(cluster, namespace, releaseName);
      return;
    }
    if (error || !app) {
      return;
    }

    let manifest: IResource[] = yaml
      .parseAllDocuments(app.manifest)
      .map((doc: yaml.ast.Document) => doc.toJSON());
    // Filter out elements in the manifest that does not comply
    // with { kind: foo }
    manifest = manifest.filter(r => r && r.kind);
    if (!isEqual(manifest, this.state.manifest)) {
      this.setState({ manifest });
    } else {
      return;
    }

    // Iterate over the current manifest to populate the initial state
    this.setState(this.parseResources(manifest, app.namespace));
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

    return this.props.app && this.props.app.info ? (
      this.appInfo(this.props.app, this.props.app.info)
    ) : (
      <LoadingWrapper />
    );
  }

  public appInfo(app: IRelease, info: hapi.release.IInfo) {
    const { cluster, push } = this.props;
    const {
      serviceRefs,
      ingressRefs,
      deployRefs,
      statefulSetRefs,
      daemonSetRefs,
      secretRefs,
      otherResources,
    } = this.state;
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
                <ChartInfo app={app} cluster={cluster} />
              </div>
              <div className="col-9">
                <div className="row padding-t-bigger">
                  <div className="col-4">
                    <ApplicationStatus
                      deployRefs={deployRefs}
                      statefulsetRefs={statefulSetRefs}
                      daemonsetRefs={daemonSetRefs}
                      info={info}
                    />
                  </div>
                  <div className="col-8 text-r">
                    <AppControls
                      cluster={cluster}
                      app={app}
                      deleteApp={this.deleteApp}
                      push={push}
                    />
                  </div>
                </div>
                <AccessURLTable serviceRefs={serviceRefs} ingressRefs={ingressRefs} />
                <AppNotes notes={app.info && app.info.status && app.info.status.notes} />
                <ResourceTable resourceRefs={secretRefs} title="Secrets" />
                <ResourceTable resourceRefs={deployRefs} title="Deployments" />
                <ResourceTable resourceRefs={statefulSetRefs} title="StatefulSets" />
                <ResourceTable resourceRefs={daemonSetRefs} title="DaemonSets" />
                <ResourceTable resourceRefs={serviceRefs} title="Services" />
                <ResourceTable resourceRefs={otherResources} title="Other Resources" />
                <AppValues values={(app.config && app.config.raw) || ""} />
              </div>
            </div>
          </div>
        </main>
      </section>
    );
  }

  private parseResources(
    resources: Array<IResource | IK8sList<IResource, {}>>,
    releaseNamespace: string,
  ): IPartialAppViewState {
    const result: IPartialAppViewState = {
      ingressRefs: [],
      deployRefs: [],
      statefulSetRefs: [],
      daemonSetRefs: [],
      otherResources: [],
      serviceRefs: [],
      secretRefs: [],
    };
    resources.forEach(i => {
      // The item may be a list
      const itemList = i as IK8sList<IResource, {}>;
      if (itemList.items) {
        // If the resource  has a list of items, treat them as a list
        // A List can contain an arbitrary set of resources so we treat them as an
        // additional manifest. We merge the current result with the resources of
        // the List, concatenating items from both.
        assignWith(
          result,
          this.parseResources((i as IK8sList<IResource, {}>).items, releaseNamespace),
          // Merge the list with the current result
          (prev, newArray) => prev.concat(newArray),
        );
      } else {
        const item = i as IResource;
        const resource = { isFetching: true, item };
        switch (i.kind) {
          case "Deployment":
            result.deployRefs.push(
              new ResourceRef(resource.item, this.props.cluster, releaseNamespace),
            );
            break;
          case "StatefulSet":
            result.statefulSetRefs.push(
              new ResourceRef(resource.item, this.props.cluster, releaseNamespace),
            );
            break;
          case "DaemonSet":
            result.daemonSetRefs.push(
              new ResourceRef(resource.item, this.props.cluster, releaseNamespace),
            );
            break;
          case "Service":
            result.serviceRefs.push(
              new ResourceRef(resource.item, this.props.cluster, releaseNamespace),
            );
            break;
          case "Ingress":
            result.ingressRefs.push(
              new ResourceRef(resource.item, this.props.cluster, releaseNamespace),
            );
            break;
          case "Secret":
            result.secretRefs.push(
              new ResourceRef(resource.item, this.props.cluster, releaseNamespace),
            );
            break;
          default:
            result.otherResources.push(
              new ResourceRef(resource.item, this.props.cluster, releaseNamespace),
            );
        }
      }
    });
    return result;
  }

  private deleteApp = (purge: boolean) => {
    return this.props.deleteApp(
      this.props.cluster,
      this.props.namespace,
      this.props.releaseName,
      purge,
    );
  };
}

export default AppView;
