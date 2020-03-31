import * as React from "react";
import { AlertTriangle } from "react-feather";

import { IK8sList, IKubeItem, IResource, ISecret } from "../../../../shared/types";
import LoadingWrapper, { LoaderType } from "../../../LoadingWrapper";
import DaemonSetItemRow from "./DaemonSetItem";
import DeploymentItemRow from "./DeploymentItem";
import OtherResourceItem from "./OtherResourceItem";
import SecretItem from "./SecretItem/SecretItem";
import ServiceItem from "./ServiceItem/ServiceItem";
import StatefulSetItemRow from "./StatefulSetItem";

interface IResourceItemProps {
  name: string;
  resource?: IKubeItem<IResource | ISecret | IK8sList<IResource | ISecret, {}>>;
  watchResource: () => void;
  closeWatch: () => void;
  avoidEmptyResouce?: boolean;
}

class WorkloadItem extends React.Component<IResourceItemProps> {
  public componentDidMount() {
    this.props.watchResource();
  }

  public componentWillUnmount() {
    this.props.closeWatch();
  }

  public render() {
    const { resource } = this.props;
    return this.renderInfo(resource);
  }

  private renderInfo(
    resource?: IKubeItem<IResource | ISecret | IK8sList<IResource | ISecret, {}>>,
  ) {
    const { name } = this.props;
    if (resource === undefined || resource.isFetching) {
      return (
        <tr className="flex">
          <td className="col-3">{name}</td>
          <td className="col-9">
            <LoadingWrapper type={LoaderType.Placeholder} />
          </td>
        </tr>
      );
    }
    if (resource.error) {
      return (
        <tr className="flex">
          {" "}
          <td className="col-3">{name}</td>
          <td className="col-9">
            <span className="flex">
              <AlertTriangle />
              <span className="flex margin-l-normal">Error: {resource.error.message}</span>
            </span>
          </td>
        </tr>
      );
    }
    if (resource.item) {
      const listItem = resource.item as IK8sList<IResource | ISecret, {}>;
      if (listItem.items) {
        if (listItem.items.length === 0) {
          if (this.props.avoidEmptyResouce) {
            return null;
          }
          return (
            <tr className="flex">
              <td className="col-12">No resource found</td>
            </tr>
          );
        }
        return listItem.items.map(i => (
          <tr key={i.metadata.selfLink} className="flex">
            {this.renderResource(i)}
          </tr>
        ));
      }
      return <tr className="flex">{this.renderResource(resource.item as IResource)}</tr>;
    }
    return <span>No resource found</span>;
  }

  private renderResource = (r: IResource | ISecret) => {
    const plainResource = r as IResource;
    // resource kind may not be available for Lists
    if (r.metadata.selfLink.match("deployments")) {
      return <DeploymentItemRow resource={plainResource} />;
    } else if (r.metadata.selfLink.match("statefulsets")) {
      return <StatefulSetItemRow resource={plainResource} />;
    } else if (r.metadata.selfLink.match("daemonsets")) {
      return <DaemonSetItemRow resource={plainResource} />;
    } else if (r.metadata.selfLink.match("services")) {
      return <ServiceItem resource={plainResource} />;
    } else if (r.metadata.selfLink.match("secrets")) {
      return <SecretItem resource={r as ISecret} />;
    } else {
      return <OtherResourceItem resource={plainResource} />;
    }
  };
}

export default WorkloadItem;
