import * as crypto from "crypto";
import * as React from "react";

import { Auth } from "../../shared/Auth";
import { IDeploymentStatus, IFunction, IResource } from "../../shared/types";
import WebSocketHelper from "../../shared/WebSocketHelper";
import DeploymentStatus from "../DeploymentStatus";
import FunctionControls from "./FunctionControls";
import FunctionEditor from "./FunctionEditor";
import FunctionInfo from "./FunctionInfo";
import FunctionLogs from "./FunctionLogs";
import FunctionTester from "./FunctionTester";

interface IFunctionViewProps {
  name: string;
  namespace: string;
  function: IFunction | undefined;
  podName: string | undefined;
  getFunction: () => Promise<void>;
  getPodName: (fn: IFunction) => Promise<string>;
  deleteFunction: () => Promise<void>;
  updateFunction: (fn: Partial<IFunction>) => Promise<void>;
}

interface IFunctionViewState {
  deployment?: IResource;
  deploymentHealthy: boolean;
  socket?: WebSocket;
  functionCode: string;
  codeModified: boolean;
}

class FunctionView extends React.Component<IFunctionViewProps, IFunctionViewState> {
  public state: IFunctionViewState = {
    codeModified: false,
    deploymentHealthy: false,
    functionCode: "",
  };

  public async componentDidMount() {
    const { getFunction } = this.props;
    getFunction();
  }

  public componentWillReceiveProps(nextProps: IFunctionViewProps) {
    if (this.state.deployment || !nextProps.function) {
      // receiving updated function after saving
      this.setState({ codeModified: false });
      return;
    }

    const f = nextProps.function;
    const apiBase = WebSocketHelper.apiBase();
    const socket = new WebSocket(
      `${apiBase}/apis/apps/v1beta1/namespaces/${
        f.metadata.namespace
      }/deployments?watch=true&labelSelector=function=${f.metadata.name}`,
      Auth.wsProtocols(),
    );
    socket.addEventListener("message", e => this.handleEvent(e));
    this.setState({
      functionCode: f.spec.function,
      socket,
    });
  }

  public componentWillUnmount() {
    const { socket } = this.state;
    if (socket) {
      socket.close();
    }
  }

  public handleEvent(e: MessageEvent) {
    const msg = JSON.parse(e.data);
    const deployment: IResource = msg.object;
    this.setState({ deployment });
    // refetch pod name when deployment is available, in case in changed
    // TODO: move deployment status into redux store
    const status: IDeploymentStatus = deployment.status;
    if (!status.availableReplicas || status.availableReplicas !== status.replicas) {
      this.setState({ deploymentHealthy: false });
    } else {
      if (!this.state.deploymentHealthy && this.props.function) {
        this.props.getPodName(this.props.function);
      }
      this.setState({ deploymentHealthy: true });
    }
  }

  public render() {
    const { function: f, deleteFunction, podName } = this.props;
    const { deployment } = this.state;
    if (!f || !deployment) {
      return <div>Loading</div>;
    }
    return (
      <section className="FunctionView padding-b-big">
        <main>
          <div className="container">
            <div className="row collapse-b-tablet">
              <div className="col-3">
                <FunctionInfo function={f} />
              </div>
              <div className="col-9">
                <div className="row padding-t-bigger">
                  <div className="col-4">
                    <DeploymentStatus deployments={[deployment]} />
                  </div>
                  <div className="col-8 text-r">
                    <FunctionControls
                      enableSaveButton={this.state.codeModified}
                      updateFunction={this.handleFunctionUpdate}
                      deleteFunction={deleteFunction}
                      namespace={f.metadata.namespace}
                    />
                  </div>
                </div>
                <FunctionEditor
                  runtime={f.spec.runtime}
                  value={this.state.functionCode}
                  onChange={this.handleCodeChange}
                />
                <div className="row" style={{ margin: "0" }}>
                  <div className="col-6">
                    <FunctionTester function={f} />
                  </div>
                  <div className="col-6">
                    <FunctionLogs function={f} podName={podName} />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </main>
      </section>
    );
  }

  private handleCodeChange = (value: string) => {
    this.setState({ functionCode: value, codeModified: true });
  };

  private handleFunctionUpdate = () => {
    const { function: f } = this.props;
    if (f) {
      this.props.updateFunction({
        ...f,
        spec: {
          ...f.spec,
          checksum: `sha256:${crypto
            .createHash("sha256")
            .update(this.state.functionCode, "utf8")
            .digest()
            .toString("hex")}`,
          function: this.state.functionCode,
        },
      });
    }
  };
}

export default FunctionView;
