import * as React from "react";

import LoadingWrapper, { LoaderType } from "../../../components/LoadingWrapper";
import { IKubeItem, IResource, IServiceSpec } from "../../../shared/types";
import isSomeResourceLoading from "../helpers";
import AccessURLItem from "./AccessURLItem";
import { GetURLItemFromIngress } from "./AccessURLItem/AccessURLIngressHelper";
import { GetURLItemFromService } from "./AccessURLItem/AccessURLServiceHelper";

interface IAccessURLTableProps {
  services: Array<IKubeItem<IResource>>;
  ingresses: Array<IKubeItem<IResource>>;
  fetchIngresses: () => void;
}

class AccessURLTable extends React.Component<IAccessURLTableProps> {
  public componentDidMount() {
    // Fetch all related Ingress resources. We don't need to fetch Services as
    // they are expected to be watched by the ServiceTable.
    this.props.fetchIngresses();
  }

  public render() {
    const { ingresses, services } = this.props;
    return (
      <React.Fragment>
        <h6>Access URLs</h6>
        <LoadingWrapper
          loaded={!isSomeResourceLoading(ingresses.concat(services))}
          type={LoaderType.Placeholder}
        >
          {this.accessTableSection()}
        </LoadingWrapper>
      </React.Fragment>
    );
  }

  private publicServices(): Array<IKubeItem<IResource>> {
    const { services } = this.props;
    const publicServices: Array<IKubeItem<IResource>> = [];
    services.forEach(s => {
      if (s.item) {
        const spec = s.item.spec as IServiceSpec;
        if (spec.type === "LoadBalancer") {
          publicServices.push(s);
        }
      }
    });
    return publicServices;
  }

  private accessTableSection() {
    const { ingresses } = this.props;
    let accessTableSection = <p>The current application does not expose a public URL.</p>;
    const publicServices = this.publicServices();
    if (publicServices.length > 0 || ingresses.length > 0) {
      accessTableSection = (
        <React.Fragment>
          <table>
            <thead>
              <tr>
                <th>RESOURCE</th>
                <th>TYPE</th>
                <th>URL</th>
              </tr>
            </thead>
            <tbody>
              {ingresses.map(i => this.renderTableEntry(i))}
              {publicServices.map(s => this.renderTableEntry(s))}
            </tbody>
          </table>
        </React.Fragment>
      );
    }
    return accessTableSection;
  }

  private renderTableEntry(i: IKubeItem<IResource>) {
    if (i.error) {
      return (
        <tr key={i.error.message}>
          <td colSpan={3}>Error: {i.error.message}</td>
        </tr>
      );
    }
    if (i.item) {
      const urlItem =
        i.item.kind === "Ingress" ? GetURLItemFromIngress(i.item) : GetURLItemFromService(i.item);
      return <AccessURLItem key={`accessURL/${i.item.metadata.name}`} URLItem={urlItem} />;
    }
    return;
  }
}

export default AccessURLTable;
