import * as React from "react";

import LoadingWrapper from "../../../components/LoadingWrapper";
import { IKubeItem, IResource } from "../../../shared/types";
import isSomeResourceLoading from "../helpers";
import ServiceItem from "./ServiceItem";

interface IServiceTableProps {
  services: Array<IKubeItem<IResource>>;
}

class ServiceTable extends React.Component<IServiceTableProps> {
  public render() {
    const { services } = this.props;
    return (
      <React.Fragment>
        <h6>Services</h6>
        <LoadingWrapper loaded={!isSomeResourceLoading(services)} size="small">
          {this.serviceSection()}
        </LoadingWrapper>
      </React.Fragment>
    );
  }

  private serviceSection() {
    const { services } = this.props;
    let serviceSection = <p>The current application does not contain any service.</p>;
    if (services.length > 0) {
      serviceSection = (
        <table>
          <thead>
            <tr>
              <th>NAME</th>
              <th>TYPE</th>
              <th>CLUSTER-IP</th>
              <th>EXTERNAL-IP</th>
              <th>PORT(S)</th>
            </tr>
          </thead>
          <tbody>
            {services.map(
              s =>
                s.item && <ServiceItem key={`services/${s.item.metadata.name}`} service={s.item} />,
            )}
          </tbody>
        </table>
      );
    }
    return serviceSection;
  }
}

export default ServiceTable;
