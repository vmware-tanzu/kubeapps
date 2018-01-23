import * as React from "react";

import { IChartAttributes } from "../../shared/types";

interface IChartMaintainersProps {
  maintainers: IChartAttributes["maintainers"];
  githubIDAsNames?: boolean;
}

class ChartMaintainers extends React.Component<IChartMaintainersProps> {
  public render() {
    const maintainerLinks = this.props.maintainers.map((v, i) => {
      let link: string | JSX.Element = v.name;
      if (this.props.githubIDAsNames) {
        link = (
          <a href={`https://github.com/${v.name}`} target="_blank">
            {v.name}
          </a>
        );
      } else if (v.email) {
        link = <a href={`mailto:${v.email}`}>{v.name}</a>;
      }
      return <li key={i}>{link}</li>;
    });
    return (
      <div className="ChartMaintainers">
        <ul className="remove-style padding-l-reset margin-b-reset">{maintainerLinks}</ul>
      </div>
    );
  }
}

export default ChartMaintainers;
