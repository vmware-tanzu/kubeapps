import * as React from "react";
import { IServiceBroker } from "../../shared/ServiceCatalog";

interface ISyncButtonProps {
  broker: IServiceBroker;
  sync: (broker: IServiceBroker) => Promise<{}>;
}

interface ISyncButtonState {
  broker: IServiceBroker | undefined;
  isSyncing: boolean;
  error: string | undefined;
}

class SyncButton extends React.Component<ISyncButtonProps, ISyncButtonState> {
  public state: ISyncButtonState = {
    broker: this.props.broker,
    error: undefined,
    isSyncing: false,
  };

  public handleSync = async () => {
    const { sync, broker } = this.props;
    this.setState({ isSyncing: true });

    try {
      await sync(broker).then(async () => this.setState({ isSyncing: false }));
    } catch (err) {
      this.setState({ isSyncing: false, error: err.toString() });
    }
  };

  public render() {
    return (
      <div className="SyncButton">
        {this.state.isSyncing && <div>Syncing...</div>}
        <button
          className="button button-primary"
          disabled={this.state.isSyncing}
          onClick={this.handleSync}
        >
          Sync
        </button>
      </div>
    );
  }
}

export default SyncButton;
