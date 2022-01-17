import { CdsControlMessage } from "@cds/react/forms";
import { CdsInput } from "@cds/react/input";
import { CdsSelect } from "@cds/react/select";
import actions from "actions";
import Tooltip from "components/js/Tooltip";
import { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import "./AppRepoAddDockerCreds.css";

interface IAppRepoFormProps {
  imagePullSecrets: string[];
  selectPullSecret: (imagePullSecret: string) => void;
  selectedImagePullSecret: string;
  namespace: string;
  appVersion: string;
  required: boolean;
  disabled: boolean;
}

export function AppRepoAddDockerCreds({
  imagePullSecrets,
  selectPullSecret,
  selectedImagePullSecret,
  namespace,
  appVersion,
  required,
  disabled,
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

  useEffect(() => {
    if (imagePullSecrets.length && !currentImagePullSecrets.length) {
      setCurrentImagePullSecrets(imagePullSecrets);
    }
  }, [imagePullSecrets, currentImagePullSecrets]);

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
      setCurrentImagePullSecrets(currentImagePullSecrets.concat(secretName));
      setUser("");
      setSecretName("");
      setPassword("");
      setEmail("");
      setServer("");
      setShowSecretSubForm(false);
    }
  };
  /* eslint-disable jsx-a11y/label-has-associated-control */
  return (
    <>
      <CdsSelect>
        <label>
          Associate Docker Registry Credentials{required ? "" : " (optional)"}{" "}
          <span className="tooltip-wrapper">
            <Tooltip
              label="pending-tooltip"
              id={`pending-tooltip`}
              icon="help"
              position="bottom-right"
              large={true}
              iconProps={{ solid: true, size: "sm" }}
            >
              You can only associate Docker Registry Credentials to namespaced Application
              Repositories (non-global). More info{" "}
              <a
                target="_blank"
                rel="noreferrer"
                href="https://github.com/kubeapps/kubeapps/blob/main/docs/user/private-app-repository.md#associating-docker-image-pull-secrets-to-an-apprepository"
              >
                here
              </a>
              .
            </Tooltip>
          </span>
        </label>
        {currentImagePullSecrets.length ? (
          <CdsControlMessage>
            Select existing secret(s) to access a private Docker registry and pull images from it.
            More info{" "}
            <a
              href={`https://github.com/kubeapps/kubeapps/blob/${appVersion}/docs/user/private-app-repository.md`}
              target="_blank"
              rel="noopener noreferrer"
            >
              here
            </a>
            .
          </CdsControlMessage>
        ) : (
          <CdsControlMessage>No existing credentials found.</CdsControlMessage>
        )}
        <select
          value={selectedImagePullSecret}
          required={required}
          onChange={e => selectPullSecret(e.target.value)}
          disabled={disabled}
        >
          <option />
          {currentImagePullSecrets.map(secretName => {
            return <option key={`option-${secretName}`}>{secretName}</option>;
          })}
        </select>
      </CdsSelect>
      {showSecretSubForm && (
        <div className="docker-creds-subform">
          <h6>New Docker Registry Credentials</h6>
          <CdsInput className="margin-t-sm">
            <label>Secret Name</label>
            <input
              id="kubeapps-docker-cred-secret-name"
              value={secretName}
              onChange={handleSecretNameChange}
              placeholder="Secret"
              required={true}
            />
          </CdsInput>
          <CdsInput className="margin-t-sm">
            <label>Server</label>
            <input
              id="kubeapps-docker-cred-server"
              value={server}
              onChange={handleServerChange}
              placeholder="https://index.docker.io/v1/"
              required={true}
            />
          </CdsInput>
          <CdsInput className="margin-t-sm">
            <label>Username</label>
            <input
              id="kubeapps-docker-cred-username"
              value={user}
              onChange={handleUserChange}
              placeholder="Username"
              required={true}
            />
          </CdsInput>
          <CdsInput className="margin-t-sm">
            <label>Password</label>
            <input
              id="kubeapps-docker-cred-password"
              type="password"
              value={password}
              onChange={handlePasswordChange}
              placeholder="Password"
              required={true}
            />
          </CdsInput>
          <CdsInput className="margin-t-sm">
            <label>Email</label>
            <input
              id="kubeapps-docker-cred-email"
              value={email}
              onChange={handleEmailChange}
              placeholder="user@example.com"
            />
          </CdsInput>
          {/* TODO(andresmgot): CdsButton "type" property doesn't work, so we need to use a normal <button>
                https://github.com/vmware/clarity/issues/5038
                */}
          <div className="margin-t-sm">
            <button
              className="btn btn-info-outline"
              type="button"
              disabled={creating}
              onClick={handleInstallClick}
            >
              {creating ? "Creating..." : "Submit"}
            </button>
            <button
              className="btn btn-info-outline"
              type="button"
              disabled={creating}
              onClick={toggleCredSubForm}
            >
              Cancel
            </button>
          </div>
        </div>
      )}
      <div hidden={showSecretSubForm} className="docker-creds-subform-button">
        <button
          className="btn btn-info-outline"
          type="button"
          disabled={disabled || creating}
          onClick={toggleCredSubForm}
        >
          Add new credentials
        </button>
      </div>
    </>
  );
}

export default AppRepoAddDockerCreds;
