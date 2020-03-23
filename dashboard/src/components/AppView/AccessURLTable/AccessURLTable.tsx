import * as React from "react";

import ResourceRef from "shared/ResourceRef";
import LoadingWrapper, { LoaderType } from "../../../components/LoadingWrapper";
import { IK8sList, IKubeItem, IResource, IServiceSpec } from "../../../shared/types";
import isSomeResourceLoading from "../helpers";
import AccessURLItem from "./AccessURLItem";
import { GetURLItemFromIngress } from "./AccessURLItem/AccessURLIngressHelper";
import { GetURLItemFromService } from "./AccessURLItem/AccessURLServiceHelper";

interface IAccessURLTableProps {
  services: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>;
  ingresses: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>;
  ingressRefs: ResourceRef[];
  getResource: (r: ResourceRef) => void;
}

class AccessURLTable extends React.Component<IAccessURLTableProps> {
  public componentDidMount() {
    this.fetchIngresses();
  }

  public componentDidUpdate(prevProps: IAccessURLTableProps) {
    if (prevProps.ingressRefs.length !== this.props.ingressRefs.length) {
      this.fetchIngresses();
    }
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
        const listItem = s.item as IK8sList<IResource, {}>;
        if (listItem.items) {
          listItem.items.forEach(item => {
            if (item.spec.type === "LoadBalancer") {
              publicServices.push({ isFetching: false, item });
            }
          });
        } else {
          const spec = (s.item as IResource).spec as IServiceSpec;
          if (spec.type === "LoadBalancer") {
            publicServices.push(s as IKubeItem<IResource>);
          }
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

  private renderTableEntry(i: IKubeItem<IResource | IK8sList<IResource, {}>>) {
    if (i.error) {
      return (
        <tr key={i.error.message}>
          <td colSpan={3}>Error: {i.error.message}</td>
        </tr>
      );
    }
    if (i.item) {
      const listItem = i.item as IK8sList<IResource, {}>;
      if (listItem.items) {
        return listItem.items.map(item => this.renderURLItem(item));
      } else {
        return this.renderURLItem(i.item as IResource);
      }
    }
    return;
  }

  private renderURLItem(i: IResource) {
    const urlItem = i.kind === "Ingress" ? GetURLItemFromIngress(i) : GetURLItemFromService(i);
    return <AccessURLItem key={`accessURL/${i.metadata.name}`} URLItem={urlItem} />;
  }

  private fetchIngresses() {
    // Fetch all related Ingress resources. We don't need to fetch Services as
    // they are expected to be watched by the ServiceTable.
    this.props.ingressRefs.forEach(r => this.props.getResource(r));
  }
}

export default AccessURLTable;
