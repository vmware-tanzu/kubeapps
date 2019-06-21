import * as React from "react";
import { AlertTriangle } from "react-feather";

import { IKubeItem, IResource, ISecret } from "../../../../shared/types";
import LoadingWrapper, { LoaderType } from "../../../LoadingWrapper";
import DaemonSetItemRow from "./DaemonSetItem";
import DeploymentItemRow from "./DeploymentItem";
import SecretItem from "./SecretItem/SecretItem";
import ServiceItem from "./ServiceItem/ServiceItem";
import StatefulSetItemRow from "./StatefulSetItem";

interface IResourceItemProps {
  name: string;
  resource?: IKubeItem<IResource>;
  watchResource: () => void;
  closeWatch: () => void;
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
    return <tr className="flex">{this.renderInfo(resource)}</tr>;
  }

  private renderInfo(resource?: IKubeItem<IResource>) {
    const { name } = this.props;
    if (resource === undefined || resource.isFetching) {
      return (
        <React.Fragment>
          <td className="col-3">{name}</td>
          <td className="col-9">
            <LoadingWrapper type={LoaderType.Placeholder} />
          </td>
        </React.Fragment>
      );
    }
    if (resource.error) {
      return (
        <React.Fragment>
          <td className="col-3">{name}</td>
          <td className="col-9">
            <span className="flex">
              <AlertTriangle />
              <span className="flex margin-l-normal">Error: {resource.error.message}</span>
            </span>
          </td>
        </React.Fragment>
      );
    }
    if (resource.item) {
      switch (resource.item.kind) {
        case "Deployment":
          return <DeploymentItemRow resource={resource.item} />;
        case "StatefulSet":
          return <StatefulSetItemRow resource={resource.item} />;
        case "DaemonSet":
          return <DaemonSetItemRow resource={resource.item} />;
        case "Service":
          return <ServiceItem resource={resource.item} />;
        case "Secret":
          // TODO: Set double type
          const i = (resource.item as any) as ISecret;
          return <SecretItem resource={i} />;
      }
    }
    return null;
  }
}

export default WorkloadItem;
