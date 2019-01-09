import * as React from "react";

import ServiceItem from "../../../containers/ServiceItemContainer";
import { IResourceRef } from "../../../shared/types";

interface IServiceTableProps {
  serviceRefs: IResourceRef[];
}

class ServiceTable extends React.Component<IServiceTableProps> {
  public render() {
    return (
      <React.Fragment>
        <h6>Services</h6>
        {this.serviceSection()}
      </React.Fragment>
    );
  }

  private serviceSection() {
    const { serviceRefs } = this.props;
    let serviceSection = <p>The current application does not contain any Service objects.</p>;
    if (serviceRefs.length > 0) {
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
            {serviceRefs.map(s => (
              <ServiceItem key={`services/${s.namespace}/${s.name}`} serviceRef={s} />
            ))}
          </tbody>
        </table>
      );
    }
    return serviceSection;
  }
}

export default ServiceTable;
