import * as React from "react";
import * as Modal from "react-modal";
import { Redirect } from "react-router";

interface IAppRepoFormProps {
  name: string;
  url: string;
  message?: string;
  redirectTo?: string;
  install: (name: string, url: string) => Promise<any>;
  update: (values: { name?: string; url?: string }) => void;
  onAfterInstall?: () => Promise<any>;
}

export const AppRepoForm = (props: IAppRepoFormProps) => {
  const { name, url, update, install, onAfterInstall } = props;
  const handleInstallClick = async () => {
    await install(name, url);
    if (onAfterInstall) {
      await onAfterInstall();
    }
  };
  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) =>
    update({ name: e.target.value });
  const handleURLChange = (e: React.ChangeEvent<HTMLInputElement>) =>
    update({ url: e.target.value });
  return (
    <div className="app-repo-form">
      <h1>Add an App Repository</h1>
      <label>
        Name:
        <input type="text" value={name} onChange={handleNameChange} />
      </label>
      <label>
        URL:
        <input type="text" value={url} onChange={handleURLChange} />
      </label>
      <button className="button button-primary" onClick={handleInstallClick}>
        Install Repo
      </button>
      {props.redirectTo && <Redirect to={props.redirectTo} />}
    </div>
  );
};

interface IAppRepoAddButtonProps {
  install: (name: string, url: string) => Promise<any>;
  redirectTo?: string;
}
interface IAppRepoAddButtonState {
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
    error: undefined,
    modalIsOpen: false,
    name: "",
    url: "",
  };

  public render() {
    const { redirectTo, install } = this.props;
    const { name, url } = this.state;
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
            <div className="container padding-v-bigger bg-action">{this.state.error}</div>
          )}
          <AppRepoForm
            name={name}
            url={url}
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
  private updateValues = async (values: { name: string; url: string }) =>
    this.setState({ ...values });
}
