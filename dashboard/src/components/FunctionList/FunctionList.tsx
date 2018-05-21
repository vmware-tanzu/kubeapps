import * as React from "react";

import { ForbiddenError, IFunction, IRBACRole, IRuntime } from "../../shared/types";
import { escapeRegExp } from "../../shared/utils";
import { CardGrid } from "../Card";
import { MessageAlert, PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";
import PageHeader from "../PageHeader";
import SearchFilter from "../SearchFilter";
import FunctionDeployButton from "./FunctionDeployButton";
import FunctionListItem from "./FunctionListItem";

interface IFunctionListProps {
  filter: string;
  functions: IFunction[];
  runtimes: IRuntime[];
  fetchRuntimes: () => Promise<any>;
  createError: Error;
  error: Error;
  fetchFunctions: (namespace: string) => Promise<any>;
  deployFunction: (n: string, ns: string, spec: IFunction["spec"]) => Promise<boolean>;
  namespace: string;
  navigateToFunction: (n: string, ns: string) => any;
  pushSearchFilter: (filter: string) => any;
}

interface IFunctionListState {
  filter: string;
}

const RequiredRBACRoles: IRBACRole[] = [
  {
    apiGroup: "kubeless.io",
    resource: "functions",
    verbs: ["list"],
  },
  {
    apiGroup: "",
    namespace: "kubeless",
    resource: "configmaps/kubeless-config",
    verbs: ["get"],
  },
];

class FunctionList extends React.Component<IFunctionListProps, IFunctionListState> {
  public state: IFunctionListState = { filter: "" };
  public componentDidMount() {
    const { filter, fetchFunctions, fetchRuntimes, namespace } = this.props;
    fetchFunctions(namespace);
    fetchRuntimes();
    this.setState({ filter });
  }

  public componentWillReceiveProps(nextProps: IFunctionListProps) {
    const { error, filter, fetchFunctions, namespace } = this.props;
    // refetch if new namespace or error removed due to location change
    if (nextProps.namespace !== namespace || (error && !nextProps.error)) {
      fetchFunctions(nextProps.namespace);
    }
    if (nextProps.filter !== filter) {
      this.setState({ filter: nextProps.filter });
    }
  }

  public render() {
    const { functions, pushSearchFilter } = this.props;
    const functionItems = this.filteredFunctions(functions, this.state.filter).map(f => (
      <FunctionListItem key={`${f.metadata.namespace}/${f.metadata.name}`} function={f} />
    ));
    return (
      <section className="FunctionList">
        <PageHeader>
          <div className="col-8">
            <div className="row collapse-b-phone-land">
              <h1>Functions</h1>
              <SearchFilter
                className="margin-l-big "
                placeholder="search functions..."
                onChange={this.handleFilterQueryChange}
                value={this.state.filter}
                onSubmit={pushSearchFilter}
              />
            </div>
          </div>
          {this.props.functions.length > 0 && (
            <div className="col-4 text-r align-center">
              <FunctionDeployButton
                error={this.props.createError}
                deployFunction={this.props.deployFunction}
                navigateToFunction={this.props.navigateToFunction}
                runtimes={this.props.runtimes}
                namespace={this.props.namespace}
              />
            </div>
          )}
        </PageHeader>
        {this.props.error ? (
          this.renderError()
        ) : this.props.functions.length === 0 ? (
          <MessageAlert header="Unleash the power of Kubeless">
            <div>
              <p className="margin-v-normal">
                Kubeless is a Kubernetes-native serverless framework that lets you deploy small bits
                of code (functions) without having to worry about the underlying infrastructure.
              </p>
              <div className="padding-t-normal padding-b-normal">
                <FunctionDeployButton
                  error={this.props.createError}
                  deployFunction={this.props.deployFunction}
                  navigateToFunction={this.props.navigateToFunction}
                  runtimes={this.props.runtimes}
                  namespace={this.props.namespace}
                />
              </div>
            </div>
          </MessageAlert>
        ) : (
          <CardGrid>{functionItems}</CardGrid>
        )}
      </section>
    );
  }

  private renderError() {
    const { error, namespace } = this.props;
    return error instanceof ForbiddenError ? (
      <PermissionsErrorAlert
        action="list Functions"
        namespace={namespace}
        roles={RequiredRBACRoles}
      />
    ) : (
      <UnexpectedErrorAlert />
    );
  }

  private filteredFunctions(functions: IFunction[], filter: string) {
    return functions.filter(f => new RegExp(escapeRegExp(filter), "i").test(f.metadata.name));
  }

  private handleFilterQueryChange = (filter: string) => {
    this.setState({
      filter,
    });
  };
}

export default FunctionList;
