import * as React from "react";
import { IServiceBinding } from "../../shared/ServiceBinding";
import { RemoveBindingButton } from "../InstanceView/RemoveBindingButton";

interface IBindingEntryProps {
  binding: IServiceBinding;
  addBinding: (bindingName: string, instanceName: string, namespace: string) => Promise<any>;
  getCatalog: () => Promise<any>;
}

interface IBindingEntryState {
  isExpanded: boolean;
}

export class BindingEntry extends React.Component<IBindingEntryProps, IBindingEntryState> {
  public state = {
    isExpanded: false,
  };

  public render() {
    const { binding } = this.props;
    const { name, namespace } = binding.metadata;

    const {
      instanceRef,
      secretName,
      secretDatabase,
      secretHost,
      secretPassword,
      secretPort,
      secretUsername,
    } = binding.spec;

    const statuses: Array<[string, string | undefined]> = [
      ["Instance", instanceRef.name],
      ["Secret", secretName],
      ["Database", secretDatabase],
      ["Host", secretHost],
      ["Password", secretPassword],
      ["Port", secretPort],
      ["Username", secretUsername],
    ];

    const condition = [...binding.status.conditions].shift();
    const currentStatus = condition ? (
      <div className="condition">
        <code>{condition.message}</code>
      </div>
    ) : (
      undefined
    );

    const rows = [
      <tr key={"row"}>
        <td>
          {namespace}/{name}
        </td>
        <td>{currentStatus}</td>
        <td style={{ display: "flex", justifyContent: "space-around" }}>
          <button className={"button button-primary button-small"} onClick={this.toggleExpand}>
            Expand/Collapse
          </button>
          <RemoveBindingButton binding={binding} />
        </td>
      </tr>,
    ];
    if (this.state.isExpanded) {
      rows.push(
        <tr key="info">
          <td colSpan={3}>
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
          </td>
        </tr>,
      );
    }
    return rows;
  }

  private toggleExpand = async () => this.setState({ isExpanded: !this.state.isExpanded });
}
