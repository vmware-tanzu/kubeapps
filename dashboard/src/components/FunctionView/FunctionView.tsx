import * as React from "react";

import { Auth } from "../../shared/Auth";
import {
  ForbiddenError,
  IDeploymentStatus,
  IFunction,
  IRBACRole,
  IResource,
  NotFoundError,
} from "../../shared/types";
import DeploymentStatus from "../DeploymentStatus";
import { NotFoundErrorAlert, PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";
import FunctionControls from "./FunctionControls";
import FunctionEditor from "./FunctionEditor";
import FunctionInfo from "./FunctionInfo";
import FunctionLogs from "./FunctionLogs";
import FunctionTester from "./FunctionTester";

interface IFunctionViewProps {
  errors: {
    delete?: Error;
    fetch?: Error;
    update?: Error;
  };
  name: string;
  namespace: string;
  function: IFunction | undefined;
  podName: string | undefined;
  getFunction: (n: string, ns: string) => Promise<void>;
  getPodName: (fn: IFunction) => Promise<void>;
  deleteFunction: (n: string, ns: string) => Promise<boolean>;
  updateFunction: (fn: Partial<IFunction>) => Promise<void>;
}

interface IFunctionViewState {
  deployment?: IResource;
  deploymentHealthy: boolean;
  socket?: WebSocket;
  functionCode: string;
  codeModified: boolean;
}

const RequiredRBACRoles: { [s: string]: IRBACRole[] } = {
  delete: [
    {
      apiGroup: "kubeless.io",
      resource: "functions",
      verbs: ["delete"],
    },
  ],
  update: [
    {
      apiGroup: "kubeless.io",
      resource: "functions",
      verbs: ["update"],
    },
  ],
  view: [
    {
      apiGroup: "kubeless.io",
      resource: "functions",
      verbs: ["get"],
    },
    {
      apiGroup: "apps",
      resource: "deployments",
      verbs: ["list", "watch"],
    },
    {
      apiGroup: "",
      resource: "pods",
      verbs: ["list"],
    },
    {
      apiGroup: "",
      resource: "pods/logs",
      verbs: ["get"],
    },
    {
      apiGroup: "",
      resource: "services/proxy",
      verbs: ["get", "create"],
    },
  ],
};

class FunctionView extends React.Component<IFunctionViewProps, IFunctionViewState> {
  public state: IFunctionViewState = {
    codeModified: false,
    deploymentHealthy: false,
    functionCode: "",
  };

  public componentDidMount() {
    const { getFunction, name, namespace } = this.props;
    getFunction(name, namespace);
  }

  public componentWillReceiveProps(nextProps: IFunctionViewProps) {
    const { getFunction, name, namespace } = this.props;
    if (nextProps.namespace !== namespace) {
      getFunction(name, nextProps.namespace);
      return;
    }

    if (this.state.deployment || !nextProps.function) {
      // receiving updated function after saving
      this.setState({ codeModified: false });
      return;
    }

    const f = nextProps.function;
    const apiBase = `ws://${window.location.host}/api/kube`;
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
    const { function: f, podName } = this.props;
    const { deployment } = this.state;
    if (this.props.errors.fetch) {
      return this.renderError(this.props.errors.fetch);
    }
    if (!f || !deployment) {
      return <div>Loading</div>;
    }
    return (
      <section className="FunctionView padding-b-big">
        <main>
          <div className="container">
            {this.props.errors.update && this.renderError(this.props.errors.update, "update")}
            {this.props.errors.delete && this.renderError(this.props.errors.delete, "delete")}
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
                      deleteFunction={this.handleFunctionDelete}
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

  private renderError(error: Error, action: string = "view") {
    const { namespace, name } = this.props;
    switch (error.constructor) {
      case ForbiddenError:
        return (
          <PermissionsErrorAlert
            namespace={namespace}
            roles={RequiredRBACRoles[action]}
            action={`${action} Function "${name}"`}
          />
        );
      case NotFoundError:
        return <NotFoundErrorAlert resource={`Function "${name}"`} namespace={namespace} />;
      default:
        return <UnexpectedErrorAlert />;
    }
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
          function: this.state.functionCode,
        },
      });
    }
  };

  private handleFunctionDelete = async () => {
    const { deleteFunction, name, namespace } = this.props;
    return await deleteFunction(name, namespace);
  };
}

export default FunctionView;
