import * as React from "react";

import LoadingWrapper from "../../../components/LoadingWrapper";
import { IKubeItem, IResource, IServiceSpec } from "../../../shared/types";
import isSomeResourceLoading from "../helpers";
import AccessURLItem from "./AccessURLItem";
import { GetURLItemFromIngress } from "./AccessURLItem/AccessURLIngressHelper";
import { GetURLItemFromService } from "./AccessURLItem/AccessURLServiceHelper";

interface IServiceTableProps {
  services: Array<IKubeItem<IResource>>;
  ingresses: Array<IKubeItem<IResource>>;
}

class AccessURLTable extends React.Component<IServiceTableProps> {
  public render() {
    const { ingresses, services } = this.props;
    return (
      <React.Fragment>
        <h6>Access URLs</h6>
        <LoadingWrapper loaded={!isSomeResourceLoading(ingresses.concat(services))} size="small">
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
              {ingresses.map(
                i =>
                  i.item && (
                    <AccessURLItem
                      key={`accessURL/${i.item.metadata.name}`}
                      URLItem={GetURLItemFromIngress(i.item)}
                    />
                  ),
              )}
              {publicServices.map(
                s =>
                  s.item && (
                    <AccessURLItem
                      key={`accessURL/${s.item.metadata.name}`}
                      URLItem={GetURLItemFromService(s.item)}
                    />
                  ),
              )}
            </tbody>
          </table>
        </React.Fragment>
      );
    }
    return accessTableSection;
  }
}

export default AccessURLTable;
