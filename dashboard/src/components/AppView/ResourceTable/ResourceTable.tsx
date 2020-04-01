import * as React from "react";

import ResourceRef from "shared/ResourceRef";
import ResourceItemContainer from "../../../containers/ResourceItemContainer";
import { DaemonSetColumns } from "./ResourceItem/DaemonSetItem/DaemonSetItem";
import { DeploymentColumns } from "./ResourceItem/DeploymentItem/DeploymentItem";
import OtherResourceItem, {
  OtherResourceColumns,
} from "./ResourceItem/OtherResourceItem/OtherResourceItem";
import { SecretColumns } from "./ResourceItem/SecretItem/SecretItem";
import { ServiceColumns } from "./ResourceItem/ServiceItem/ServiceItem";
import { StatefulSetColumns } from "./ResourceItem/StatefulSetItem/StatefulSetItem";

interface IResourceTableProps {
  title: string;
  resourceRefs: ResourceRef[];
  requestOtherResources?: boolean;
}

class ResourceTable extends React.Component<IResourceTableProps> {
  public render() {
    const { resourceRefs } = this.props;
    let section: JSX.Element | null = null;
    if (resourceRefs.length > 0) {
      section = (
        <React.Fragment>
          <h6>{this.props.title}</h6>
          <table>
            <thead>
              <tr className="flex">{this.getColumns()}</tr>
            </thead>
            <tbody>
              {resourceRefs.map(r => {
                switch (r.kind) {
                  case "Deployment":
                  case "StatefulSet":
                  case "DaemonSet":
                  case "Service":
                  case "Secret":
                    return <ResourceItemContainer key={r.getResourceURL()} resourceRef={r} />;
                  default:
                    return this.props.requestOtherResources ? (
                      <ResourceItemContainer
                        key={r.getResourceURL()}
                        resourceRef={r}
                        avoidEmptyResouce={true}
                      />
                    ) : (
                      // We may not know the plural of the resource so we don't get the full resource URL
                      <tr key={`otherResources/${r.kind}/${r.name}`} className="flex">
                        <OtherResourceItem
                          key={`${r.kind}/${r.namespace}/${r.name}`}
                          resource={r}
                        />
                      </tr>
                    );
                }
              })}
            </tbody>
          </table>
        </React.Fragment>
      );
    }
    return section;
  }

  private getColumns() {
    switch (this.props.resourceRefs[0].kind) {
      case "Deployment":
        return <DeploymentColumns />;
      case "StatefulSet":
        return <StatefulSetColumns />;
      case "DaemonSet":
        return <DaemonSetColumns />;
      case "Service":
        return <ServiceColumns />;
      case "Secret":
        return <SecretColumns />;
      default:
        return <OtherResourceColumns />;
    }
  }
}

export default ResourceTable;
