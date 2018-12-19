import * as React from "react";
import * as Modal from "react-modal";
import { Redirect } from "react-router";

import { IRBACRole } from "../../../shared/types";
import ErrorSelector from "../../ErrorAlert/ErrorSelector";
import "./AppRepo.css";

interface IAppRepoFormProps {
  name: string;
  url: string;
  authHeader: string;
  customCA: string;
  message?: string;
  redirectTo?: string;
  install: (name: string, url: string, authHeader: string, customCA: string) => Promise<boolean>;
  update: (values: { name?: string; url?: string; authHeader?: string; customCA?: string }) => void;
  onAfterInstall?: () => Promise<any>;
}

const RequiredRBACRoles: IRBACRole[] = [
  {
    apiGroup: "kubeapps.com",
    resource: "apprepositories",
    verbs: ["create"],
  },
  {
    apiGroup: "",
    resource: "secrets",
    verbs: ["create"],
  },
];

export const AppRepoForm = (props: IAppRepoFormProps) => {
  const { name, url, authHeader, customCA, update, install, onAfterInstall } = props;
  const handleInstallClick = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const installed = await install(name, url, authHeader, customCA);
    if (installed && onAfterInstall) {
      await onAfterInstall();
    }
  };
  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) =>
    update({ name: e.target.value });
  const handleURLChange = (e: React.ChangeEvent<HTMLInputElement>) =>
    update({ url: e.target.value });
  const handleAuthHeaderChange = (e: React.ChangeEvent<HTMLInputElement>) =>
    update({ authHeader: e.target.value });
  const handleCustomCAChange = (e: React.ChangeEvent<HTMLTextAreaElement>) =>
    update({ customCA: e.target.value });
  return (
    <form className="container padding-b-bigger" onSubmit={handleInstallClick}>
      <div className="row">
        <div className="col-12">
          <div>
            <h2>Add an App Repository</h2>
          </div>
          <div>
            <label>
              <span>Name:</span>
              <input
                type="text"
                placeholder="example"
                value={name}
                onChange={handleNameChange}
                required={true}
                pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
                title="Use lower case alphanumeric characters, '-' or '.'"
              />
            </label>
          </div>
          <div>
            <label>
              <span>URL:</span>
              <input
                type="url"
                placeholder="https://charts.example.com/stable"
                value={url}
                onChange={handleURLChange}
                required={true}
              />
            </label>
          </div>
          <div>
            <label>
              <span>Authorization Header (optional):</span>
              <input
                type="text"
                placeholder="Bearer xrxNcWghpRLdcPHFgVRM73rr4N7qjvjm"
                value={authHeader}
                onChange={handleAuthHeaderChange}
              />
            </label>
          </div>
          <div>
            <label>
              <span>Custom CA Certificate (optional):</span>
              <pre className="CertContainer">
                <textarea
                  className="CertContent"
                  rows={4}
                  placeholder={
                    "-----BEGIN CERTIFICATE-----\n" + "...\n" + "-----END CERTIFICATE-----"
                  }
                  value={customCA}
                  onChange={handleCustomCAChange}
                />
              </pre>
            </label>
          </div>
          <div>
            <button className="button button-primary" type="submit">
              Install Repo
            </button>
          </div>
          {props.redirectTo && <Redirect to={props.redirectTo} />}
        </div>
      </div>
    </form>
  );
};

interface IAppRepoAddButtonProps {
  error?: Error;
  install: (name: string, url: string, authHeader: string, customCA: string) => Promise<boolean>;
  redirectTo?: string;
  kubeappsNamespace: string;
}
interface IAppRepoAddButtonState {
  authHeader: string;
  customCA: string;
  lastSubmittedName: string;
  modalIsOpen: boolean;
  name: string;
  url: string;
}

export class AppRepoAddButton extends React.Component<
  IAppRepoAddButtonProps,
  IAppRepoAddButtonState
> {
  public state = {
    authHeader: "",
    customCA: "",
    error: undefined,
    lastSubmittedName: "",
    modalIsOpen: false,
    name: "",
    url: "",
  };

  public render() {
    const { redirectTo } = this.props;
    const { name, url, authHeader, customCA } = this.state;
    return (
      <div className="AppRepoAddButton">
        <button className="button button-primary" onClick={this.openModal}>
          Add App Repository
        </button>
        <Modal
          isOpen={this.state.modalIsOpen}
          onRequestClose={this.closeModal}
          contentLabel="Modal"
        >
          {this.props.error && (
            <ErrorSelector
              error={this.props.error}
              defaultRequiredRBACRoles={{ create: RequiredRBACRoles }}
              action="create"
              namespace={this.props.kubeappsNamespace}
              resource={`App Repository ${this.state.lastSubmittedName}`}
            />
          )}
          <AppRepoForm
            name={name}
            url={url}
            authHeader={authHeader}
            customCA={customCA}
            update={this.updateValues}
            install={this.install}
            onAfterInstall={this.closeModal}
          />
        </Modal>
        {redirectTo && <Redirect to={redirectTo} />}
      </div>
    );
  }

  private closeModal = async () => this.setState({ modalIsOpen: false });
  private install = (name: string, url: string, authHeader: string, customCA: string) => {
    // Store last submitted name to show it in an error if needed
    this.setState({ lastSubmittedName: this.state.name });
    return this.props.install(name, url, authHeader, customCA);
  };
  private openModal = async () => this.setState({ modalIsOpen: true });
  private updateValues = async (values: {
    name: string;
    url: string;
    authHeader: string;
    customCA: string;
  }) => this.setState({ ...values });
}
