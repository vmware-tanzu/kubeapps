import * as React from "react";
import * as Modal from "react-modal";
import { Redirect } from "react-router";

interface IAppRepoFormProps {
  name: string;
  url: string;
  authHeader: string;
  message?: string;
  redirectTo?: string;
  install: (name: string, url: string, authHeader: string) => Promise<any>;
  update: (values: { name?: string; url?: string; authHeader?: string }) => void;
  onAfterInstall?: () => Promise<any>;
}

export const AppRepoForm = (props: IAppRepoFormProps) => {
  const { name, url, authHeader, update, install, onAfterInstall } = props;
  const handleInstallClick = async () => {
    await install(name, url, authHeader);
    if (onAfterInstall) {
      await onAfterInstall();
    }
  };
  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) =>
    update({ name: e.target.value });
  const handleURLChange = (e: React.ChangeEvent<HTMLInputElement>) =>
    update({ url: e.target.value });
  const handleAuthHeaderChange = (e: React.ChangeEvent<HTMLInputElement>) =>
    update({ authHeader: e.target.value });
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
  install: (name: string, url: string, authHeader: string) => Promise<any>;
  redirectTo?: string;
}
interface IAppRepoAddButtonState {
  authHeader: string;
  error?: string;
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
    error: undefined,
    modalIsOpen: false,
    name: "",
    url: "",
  };

  public render() {
    const { redirectTo, install } = this.props;
    const { name, url, authHeader } = this.state;
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
          {this.state.error && (
            <div className="padding-big margin-b-big bg-action">{this.state.error}</div>
          )}
          <AppRepoForm
            name={name}
            url={url}
            authHeader={authHeader}
            update={this.updateValues}
            install={install}
            onAfterInstall={this.closeModal}
          />
        </Modal>
        {redirectTo && <Redirect to={redirectTo} />}
      </div>
    );
  }

  private closeModal = async () => this.setState({ modalIsOpen: false });
  private openModal = async () => this.setState({ modalIsOpen: true });
  private updateValues = async (values: { name: string; url: string; authHeader: string }) =>
    this.setState({ ...values });
}
