import * as crypto from "crypto";
import * as Moniker from "moniker-native";
import * as React from "react";
import AceEditor from "react-ace";
import * as Modal from "react-modal";

import "brace/mode/json";
import "brace/mode/ruby";
import "brace/mode/text";

import { IFunction, IRuntime } from "../../shared/types";

interface IFunctionDeployButtonProps {
  deployFunction: (n: string, ns: string, spec: IFunction["spec"]) => Promise<any>;
  navigateToFunction: (n: string, ns: string) => Promise<any>;
  runtimes: IRuntime[];
}

interface IFunctionDeployButtonState {
  functionSpec: IFunction["spec"];
  modalIsOpen: boolean;
  name: string;
  namespace: string;
  error?: string;
}

class FunctionDeployButton extends React.Component<
  IFunctionDeployButtonProps,
  IFunctionDeployButtonState
> {
  public state: IFunctionDeployButtonState = {
    functionSpec: {
      checksum: "",
      deps: "",
      function: "",
      handler: "hello.handler",
      runtime: "nodejs6",
    },
    modalIsOpen: false,
    name: "",
    namespace: "default",
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
    const { functionSpec: f, name, namespace } = this.state;
    const runtimes = {};
    this.props.runtimes.forEach(r => {
      r.versions.forEach(version => {
        const target = r.ID + version.version;
        runtimes[`${r.ID} (${version.version})`] = target;
      });
    });
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
          {this.state.error && (
            <div className="padding-big margin-b-big bg-action">{this.state.error}</div>
          )}
          <form onSubmit={this.handleDeploy}>
            <div>
              <label htmlFor="name">Name</label>
              <input id="name" onChange={this.handleNameChange} value={name} required={true} />
            </div>
            <div>
              <label htmlFor="namespace">Namespace</label>
              <input
                name="namespace"
                onChange={this.handleNamespaceChange}
                value={namespace}
                required={true}
              />
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
                {Object.keys(runtimes).map(r => (
                  <option key={runtimes[r]} value={runtimes[r]}>
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

  private runtimeToDepsMode() {
    const { functionSpec: { runtime } } = this.state;
    if (runtime.match(/node|php/)) {
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
    let deps = "";
    this.props.runtimes.forEach(r => {
      if (runtime.match(r.ID)) {
        deps = r.depName;
      }
    });
    return deps;
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
    const { deployFunction, navigateToFunction } = this.props;
    const { functionSpec, name, namespace } = this.state;
    const functionSha256 = crypto
      .createHash("sha256")
      .update(functionSpec.function, "utf8")
      .digest()
      .toString("hex");
    functionSpec.checksum = `sha256:${functionSha256}`;
    try {
      await deployFunction(name, namespace, functionSpec);
      navigateToFunction(name, namespace);
    } catch (err) {
      this.setState({ error: err.toString() });
    }
  };

  private handleNameChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ name: e.currentTarget.value });
  };
  private handleNamespaceChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ namespace: e.currentTarget.value });
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
  ${fnName}: function(event, context) {
    return "Hello World";
  }
};
`;
    } else if (runtime.match(/ruby/)) {
      return `def ${fnName}(event, context)
  "Hello World"
end
`;
    } else if (runtime.match(/php/)) {
      return `<?php
function ${fnName}($event, $context) {
  return "hello world";
}
`;
    } else if (runtime.match(/python/)) {
      return `def ${fnName}(event, context):
  return "Hello World"
`;
    }
    return "";
  };
}

export default FunctionDeployButton;
