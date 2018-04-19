import * as Moniker from "moniker-native";
import * as React from "react";
import AceEditor from "react-ace";
import * as Modal from "react-modal";

import "brace/mode/json";
import "brace/mode/ruby";
import "brace/mode/text";

import { ForbiddenError, IFunction, IRBACRole, NotFoundError } from "../../shared/types";
import { NotFoundErrorAlert, PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";

const RequiredRBACRoles: IRBACRole[] = [
  {
    apiGroup: "kubeless.io",
    resource: "functions",
    verbs: ["create"],
  },
];

// TODO: fetch this from the API/ConfigMap
const Runtimes = {
  "Nodejs (6)": "nodejs6",
  "Nodejs (8)": "nodejs8",
  "Python (2.7)": "python2.7",
  "Python (3.4)": "python3.4",
  "Python (3.6)": "python3.6",
  "Ruby (2.4)": "ruby2.4",
};

interface IFunctionDeployButtonProps {
  error?: Error;
  deployFunction: (n: string, ns: string, spec: IFunction["spec"]) => Promise<boolean>;
  namespace: string;
  navigateToFunction: (n: string, ns: string) => Promise<any>;
}

interface IFunctionDeployButtonState {
  functionSpec: IFunction["spec"];
  modalIsOpen: boolean;
  name: string;
  error?: string;
}

class FunctionDeployButton extends React.Component<
  IFunctionDeployButtonProps,
  IFunctionDeployButtonState
> {
  public state: IFunctionDeployButtonState = {
    functionSpec: {
      deps: "",
      function: "",
      handler: "hello.handler",
      runtime: "nodejs6",
      type: "HTTP",
    },
    modalIsOpen: false,
    name: "",
  };

  public componentDidMount() {
    const generatedName = Moniker.choose();
    this.setState({
      functionSpec: {
        ...this.state.functionSpec,
        function: this.generateFunction(
          this.state.functionSpec.runtime,
          this.state.functionSpec.handler,
        ),
        handler: `${generatedName}.handler`,
      },
      name: generatedName,
    });
  }

  public render() {
    const { functionSpec: f, name } = this.state;
    return (
      <div className="FunctionDeployButton">
        <button className="button button-accent" onClick={this.openModal}>
          Deploy New Function
        </button>
        <Modal
          isOpen={this.state.modalIsOpen}
          onRequestClose={this.closeModal}
          contentLabel="Modal"
        >
          {this.props.error && <div className="margin-b-bigger">{this.renderError()}</div>}
          <form onSubmit={this.handleDeploy}>
            <div>
              <label htmlFor="name">Name</label>
              <input id="name" onChange={this.handleNameChange} value={name} required={true} />
            </div>
            <div>
              <label htmlFor="handler">Handler</label>
              <input
                name="handler"
                onChange={this.handleHandlerChange}
                value={f.handler}
                required={true}
              />
            </div>
            <div>
              <label htmlFor="runtimes">Runtimes</label>
              <select onChange={this.handleRuntimeChange} value={f.runtime}>
                {Object.keys(Runtimes).map(r => (
                  <option key={Runtimes[r]} value={Runtimes[r]}>
                    {r}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label htmlFor="dependencies">Dependencies ({this.runtimeToDepsDescription()})</label>
              <AceEditor
                className="margin-b-bigger"
                mode={this.runtimeToDepsMode()}
                theme="xcode"
                name="dependencies"
                width="100%"
                height="200px"
                onChange={this.handleDependenciesChange}
                setOptions={{ showPrintMargin: false }}
                value={f.deps}
              />
            </div>
            <div>
              <button className="button button-primary" type="submit">
                Submit
              </button>
              <button className="button" onClick={this.closeModal}>
                Cancel
              </button>
            </div>
          </form>
        </Modal>
      </div>
    );
  }

  private renderError() {
    const { error, namespace } = this.props;
    const { name } = this.state;
    switch (error && error.constructor) {
      case ForbiddenError:
        return (
          <PermissionsErrorAlert
            namespace={namespace}
            roles={RequiredRBACRoles}
            action={`create Function "${name}"`}
          />
        );
      case NotFoundError:
        return <NotFoundErrorAlert resource={`Namespace "${namespace}"`} />;
      default:
        return <UnexpectedErrorAlert />;
    }
  }

  private runtimeToDepsMode() {
    const { functionSpec: { runtime } } = this.state;
    if (runtime.match(/node/)) {
      return "json";
    } else if (runtime.match(/ruby/)) {
      return "ruby";
    } else if (runtime.match(/python/)) {
      return "text";
    }
    return "";
  }

  private runtimeToDepsDescription() {
    const { functionSpec: { runtime } } = this.state;
    if (runtime.match(/node/)) {
      return "package.json";
    } else if (runtime.match(/ruby/)) {
      return "Gemfile";
    } else if (runtime.match(/python/)) {
      return "requirements.txt";
    }
    return "";
  }

  private openModal = () => {
    this.setState({
      modalIsOpen: true,
    });
  };

  private closeModal = () => {
    this.setState({
      modalIsOpen: false,
    });
  };

  private handleDeploy = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const { deployFunction, namespace, navigateToFunction } = this.props;
    const { functionSpec, name } = this.state;
    const created = await deployFunction(name, namespace, functionSpec);
    if (created) {
      navigateToFunction(name, namespace);
    }
  };

  private handleNameChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ name: e.currentTarget.value });
  };
  private handleHandlerChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({
      functionSpec: {
        ...this.state.functionSpec,
        function: this.generateFunction(this.state.functionSpec.runtime, e.currentTarget.value),
        handler: e.currentTarget.value,
      },
    });
  };
  private handleRuntimeChange = (e: React.FormEvent<HTMLSelectElement>) => {
    this.setState({
      functionSpec: {
        ...this.state.functionSpec,
        function: this.generateFunction(e.currentTarget.value, this.state.functionSpec.handler),
        runtime: e.currentTarget.value,
      },
    });
  };
  private handleDependenciesChange = (value: string) => {
    this.setState({ functionSpec: { ...this.state.functionSpec, deps: value } });
  };

  private generateFunction = (runtime: string, handler: string) => {
    const fnName = handler.split(".").pop();
    if (runtime.match(/node/)) {
      return `module.exports = {
  ${fnName}: function(req, res) {
    res.end("Hello World");
  }
};
`;
    } else if (runtime.match(/ruby/)) {
      return `def ${fnName}(request)
  "Hello World"
end
`;
    } else if (runtime.match(/python/)) {
      return `def ${fnName}():
  return "Hello World"
`;
    }
    return "";
  };
}

export default FunctionDeployButton;
