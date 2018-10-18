import * as React from "react";

import { IServiceBindingWithSecret } from "shared/ServiceBinding";
import "./BindingDetails.css";

class BindingDetails extends React.Component<IServiceBindingWithSecret> {
  public render() {
    const { binding, secret } = this.props;
    const { instanceRef, secretName } = binding.spec;

    const statuses: string[][] = [["Instance", instanceRef.name], ["Secret", secretName]];
    let secretDataArray: string[][] = [];
    if (secret) {
      const secretData = Object.keys(secret.data).map(k => [k, atob(secret.data[k])]);
      secretDataArray = [...secretData];
    }
    return (
      <dl className="BindingDetails container margin-normal">
        {statuses.map(statusPair => {
          const [key, value] = statusPair;
          return (
            <dt key={key}>
              {key}: <b>{value}</b>
            </dt>
          );
        })}
        {secretDataArray.map(statusPair => {
          const [key, value] = statusPair;
          return (
            <dt key={key}>
              &nbsp;&nbsp;&nbsp;&nbsp;
              {key}: <b>{value}</b>
            </dt>
          );
        })}
      </dl>
    );
  }
}

export default BindingDetails;
