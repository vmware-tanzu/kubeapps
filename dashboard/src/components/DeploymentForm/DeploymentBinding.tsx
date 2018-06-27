import * as React from "react";
import { IServiceBinding } from "../../shared/ServiceBinding";

interface IDeploymentBindingProps {
  bindings: IServiceBinding[];
  namespace: string;
  getBindings: (ns: string) => Promise<IServiceBinding[]>;
}

interface IDeploymentBindingState {
  namespace: string;
  selectedBinding: IServiceBinding | undefined;
}

class DeploymentBinding extends React.Component<IDeploymentBindingProps, IDeploymentBindingState> {
  public state: IDeploymentBindingState = {
    namespace: this.props.namespace,
    selectedBinding: undefined,
  };
  public componentDidMount() {
    this.props.getBindings(this.props.namespace);
  }

  public componentWillReceiveProps(nextProps: IDeploymentBindingProps) {
    const { namespace, getBindings } = this.props;
    if (nextProps.namespace !== namespace) {
      this.setState({ namespace: nextProps.namespace });
      getBindings(nextProps.namespace);
      return;
    }
  }

  public render() {
    const { selectedBinding } = this.state;
    if (!selectedBinding) {
      return <div />;
    } else {
      const {
        instanceRef,
        secretName,
        secretDatabase,
        secretHost,
        secretPassword,
        secretPort,
        secretUsername,
      } = selectedBinding.spec;

      const statuses: Array<[string, string | undefined]> = [
        ["Instance", instanceRef.name],
        ["Secret", secretName],
        ["Database", secretDatabase],
        ["Host", secretHost],
        ["Password", secretPassword],
        ["Port", secretPort],
        ["Username", secretUsername],
      ];

      return (
        <div>
          <p>[Optional] Select a service binding for your new app</p>
          <label htmlFor="bindings">Bindings</label>
          <select onChange={this.onBindingChange}>
            <option key="none" value="none">
              {" "}
              -- Select one --
            </option>
            {this.props.bindings.map(b => (
              <option
                key={b.metadata.name}
                selected={b.metadata.name === (selectedBinding && selectedBinding.metadata.name)}
                value={b.metadata.name}
              >
                {b.metadata.name}
              </option>
            ))}
          </select>
          <dl className="container margin-normal">
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
        </div>
      );
    }
  }

  public onBindingChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    this.setState({
      selectedBinding:
        this.props.bindings.find(binding => binding.metadata.name === e.target.value) || undefined,
    });
  };
}

export default DeploymentBinding;
