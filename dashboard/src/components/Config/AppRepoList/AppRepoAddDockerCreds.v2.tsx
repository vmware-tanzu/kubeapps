import React, { useState } from "react";

import actions from "actions";
import { CdsButton } from "components/Clarity/clarity";
import { useDispatch } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { ISecret, IStoreState } from "../../../shared/types";

interface IAppRepoFormProps {
  imagePullSecrets: ISecret[];
  togglePullSecret: (imagePullSecret: string) => () => void;
  selectedImagePullSecrets: { [key: string]: boolean };
  namespace: string;
}

export function AppRepoAddDockerCreds({
  imagePullSecrets,
  togglePullSecret,
  selectedImagePullSecrets,
  namespace,
}: IAppRepoFormProps) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const [secretName, setSecretName] = useState("");
  const [user, setUser] = useState("");
  const [password, setPassword] = useState("");
  const [email, setEmail] = useState("");
  const [server, setServer] = useState("");
  const [showSecretSubForm, setShowSecretSubForm] = useState(false);
  const [creating, setCreating] = useState(false);
  const [currentImagePullSecrets, setCurrentImagePullSecrets] = useState(imagePullSecrets);

  const handleUserChange = (e: React.ChangeEvent<HTMLInputElement>) => setUser(e.target.value);
  const handleSecretNameChange = (e: React.ChangeEvent<HTMLInputElement>) =>
    setSecretName(e.target.value);
  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) =>
    setPassword(e.target.value);
  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => setEmail(e.target.value);
  const handleServerChange = (e: React.ChangeEvent<HTMLInputElement>) => setServer(e.target.value);
  const toggleCredSubForm = () => setShowSecretSubForm(!showSecretSubForm);

  const handleInstallClick = async () => {
    setCreating(true);
    const success = await dispatch(
      actions.repos.createDockerRegistrySecret(
        secretName,
        user,
        password,
        email,
        server,
        namespace,
      ),
    );
    setCreating(false);
    if (success) {
      // Re-fetching secrets cause a re-render and the modal to be closed,
      // using local state to avoid that.
      setCurrentImagePullSecrets(
        currentImagePullSecrets.concat({ metadata: { name: secretName, namespace } } as ISecret),
      );
      setUser("");
      setSecretName("");
      setPassword("");
      setEmail("");
      setServer("");
      setShowSecretSubForm(false);
    }
  };

  return (
    <div className="clr-form-columns">
      {currentImagePullSecrets.length > 0 ? (
        currentImagePullSecrets.map(secret => {
          return (
            <div key={secret.metadata.name} className="clr-checkbox-wrapper">
              <label
                className="clr-control-label clr-control-label-checkbox"
                htmlFor={`app-repo-secret-${secret.metadata.name}`}
                key={secret.metadata.name}
              >
                <input
                  id={`app-repo-secret-${secret.metadata.name}`}
                  type="checkbox"
                  onChange={togglePullSecret(secret.metadata.name)}
                  checked={selectedImagePullSecrets[secret.metadata.name] || false}
                />
                <span>{secret.metadata.name}</span>
              </label>
            </div>
          );
        })
      ) : (
        <label className="clr-control-label">No existing credentials found.</label>
      )}
      {showSecretSubForm && (
        <div className="secondary-input">
          <label className="clr-control-label">New Docker Registry Credentials</label>
          <div className="clr-form-separator-sm">
            <label htmlFor="kubeapps-docker-cred-secret-name" className="clr-control-label">
              Secret Name
            </label>
            <div className="clr-control-container">
              <div className="clr-input-wrapper">
                <input
                  id="kubeapps-docker-cred-secret-name"
                  className="clr-input"
                  value={secretName}
                  onChange={handleSecretNameChange}
                  placeholder="Secret"
                  required={true}
                />
              </div>
            </div>
          </div>
          <div className="clr-form-control">
            <label className="clr-control-label" htmlFor="kubeapps-docker-cred-server">
              Server
            </label>
            <div className="clr-control-container">
              <div className="clr-input-wrapper">
                <input
                  id="kubeapps-docker-cred-server"
                  value={server}
                  className="clr-input"
                  onChange={handleServerChange}
                  placeholder="https://index.docker.io/v1/"
                  required={true}
                />
              </div>
            </div>
          </div>
          <div className="clr-form-control">
            <label className="clr-control-label" htmlFor="kubeapps-docker-cred-username">
              Username
            </label>
            <div className="clr-control-container">
              <div className="clr-input-wrapper">
                <input
                  id="kubeapps-docker-cred-username"
                  className="clr-input"
                  value={user}
                  onChange={handleUserChange}
                  placeholder="Username"
                  required={true}
                />
              </div>
            </div>
          </div>
          <div className="clr-form-control">
            <label className="clr-control-label" htmlFor="kubeapps-docker-cred-password">
              Password
            </label>
            <div className="clr-control-container">
              <div className="clr-input-wrapper">
                <input
                  type="password"
                  id="kubeapps-docker-cred-password"
                  className="clr-input"
                  value={password}
                  onChange={handlePasswordChange}
                  placeholder="Password"
                  required={true}
                />
              </div>
            </div>
          </div>
          <div className="clr-form-control">
            <label className="clr-control-label" htmlFor="kubeapps-docker-cred-email">
              Email
            </label>
            <div className="clr-control-container">
              <div className="clr-input-wrapper">
                <input
                  id="kubeapps-docker-cred-email"
                  className="clr-input"
                  value={email}
                  onChange={handleEmailChange}
                  placeholder="user@example.com"
                  required={true}
                />
              </div>
            </div>
          </div>
          <div className="clr-form-separator">
            <CdsButton type="button" disabled={creating} onClick={handleInstallClick}>
              {creating ? "Creating..." : "Submit"}
            </CdsButton>
            <CdsButton onClick={toggleCredSubForm} type="button" action="outline">
              Cancel
            </CdsButton>
          </div>
        </div>
      )}
      {!showSecretSubForm && (
        <div className="clr-form-separator-sm">
          <CdsButton onClick={toggleCredSubForm} type="button" size="sm">
            Add new credentials
          </CdsButton>
        </div>
      )}
    </div>
  );
}

export default AppRepoAddDockerCreds;
