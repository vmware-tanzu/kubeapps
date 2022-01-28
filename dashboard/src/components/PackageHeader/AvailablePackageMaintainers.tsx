// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { Maintainer } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import React from "react";
interface IAvailablePackageMaintainersProps {
  maintainers: Maintainer[];
  githubIDAsNames?: boolean;
}

class AvailablePackageMaintainers extends React.Component<IAvailablePackageMaintainersProps> {
  public render() {
    const maintainerLinks = this.props.maintainers.map((v, i) => {
      let link: string | JSX.Element = v.name;
      if (this.props.githubIDAsNames) {
        link = (
          <a href={`https://github.com/${v.name}`} target="_blank" rel="noopener noreferrer">
            {v.name}
          </a>
        );
      } else if (v.email) {
        link = <a href={`mailto:${v.email}`}>{v.name}</a>;
      }
      return <li key={i}>{link}</li>;
    });
    return (
      <div>
        <ul>{maintainerLinks}</ul>
      </div>
    );
  }
}

export default AvailablePackageMaintainers;
