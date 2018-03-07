import * as React from "react";

import { IFunction } from "../../shared/types";
import { CardGrid } from "../Card";
import FunctionListItem from "./FunctionListItem";

interface IFunctionListProps {
  functions: IFunction[];
  namespace: string;
  fetchFunctions: (namespace: string) => Promise<any>;
}

class FunctionList extends React.Component<IFunctionListProps> {
  public componentDidMount() {
    const { namespace, fetchFunctions } = this.props;
    fetchFunctions(namespace);
  }

  public render() {
    const chartItems = this.props.functions.map(f => (
      <FunctionListItem key={`${f.metadata.namespace}/${f.metadata.name}`} function={f} />
    ));
    return (
      <section className="FunctionList">
        <header className="FunctionList__header">
          <h1>Functions</h1>
          <hr />
        </header>
        <CardGrid>{chartItems}</CardGrid>
      </section>
    );
  }
}

export default FunctionList;
