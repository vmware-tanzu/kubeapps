import * as React from "react";
import { AlertTriangle } from "react-feather";

import LoadingWrapper, { LoaderType } from "../../../components/LoadingWrapper";
import { IKubeItem, IResource } from "../../../shared/types";

interface IResourceItemProps {
  name: string;
  resource?: IKubeItem<IResource>;
  statusFields: string[];
  watchResource: () => void;
  closeWatch: () => void;
}

class ResourceItem extends React.Component<IResourceItemProps> {
  public componentDidMount() {
    this.props.watchResource();
  }

  public componentWillUnmount() {
    this.props.closeWatch();
  }

  public render() {
    const { name, resource } = this.props;
    return (
      <tr>
        <td>{name}</td>
        {this.renderInfo(resource)}
      </tr>
    );
  }

  private renderInfo(resource?: IKubeItem<IResource>) {
    if (resource === undefined || resource.isFetching) {
      return (
        <td colSpan={this.props.statusFields.length}>
          <LoadingWrapper type={LoaderType.Placeholder} />
        </td>
      );
    }
    if (resource.error) {
      return (
        <td colSpan={this.props.statusFields.length}>
          <span className="flex">
            <AlertTriangle />
            <span className="flex margin-l-normal">Error: {resource.error.message}</span>
          </span>
        </td>
      );
    }
    if (resource.item) {
      const status = resource.item.status;
      return (
        <React.Fragment>
          {this.props.statusFields.map(c => (
            <td key={c}>{status[c] || 0}</td>
          ))}
        </React.Fragment>
      );
    }
    return null;
  }
}

export default ResourceItem;
