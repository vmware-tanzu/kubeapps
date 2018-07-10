import * as React from "react";

import { IK8sApiSecretResponse, IServiceBinding } from "shared/ServiceBinding";
import "./BindingDetails.css";

interface IBindingDetailsProps {
  binding: IServiceBinding;
  secret?: IK8sApiSecretResponse;
}

class BindingDetails extends React.Component<IBindingDetailsProps> {
  public render() {
    const { binding, secret } = this.props;
    const { instanceRef, secretName } = binding.spec;

    let statuses: string[][] = [["Instance", instanceRef.name], ["Secret", secretName]];
    if (secret) {
      const secretData = Object.keys(secret.data).map(k => [k, atob(secret.data[k])]);
      statuses = [...statuses, ...secretData];
    }
    return (
      <dl className="BindingDetails container margin-normal">
        {statuses.map(statusPair => {
          const [key, value] = statusPair;
          return [
            <dt key={key}>{key}</dt>,
            <dd key={value}>
              <code>{value}</code>
            </dd>,
          ];
        })}
      </dl>
    );
  }
}

export default BindingDetails;
