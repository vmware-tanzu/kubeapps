import * as React from "react";
import { ISecret } from "../../../shared/types";

interface IAppRepoFormProps {
  imagePullSecrets: ISecret[];
  togglePullSecret: (imagePullSecret: string) => () => void;
  selectedImagePullSecrets: { [key: string]: boolean };
}

export class AppRepoAddDockerCreds extends React.Component<IAppRepoFormProps, {}> {
  public render() {
    const { imagePullSecrets, togglePullSecret, selectedImagePullSecrets } = this.props;
    return (
      <div className="margin-l-big margin-t-normal">
        {imagePullSecrets.length > 0 ? (
          imagePullSecrets.map(secret => {
            return (
              <div key={secret.metadata.name}>
                <label
                  className="checkbox"
                  key={secret.metadata.name}
                  onChange={togglePullSecret(secret.metadata.name)}
                >
                  <input type="checkbox" checked={selectedImagePullSecrets[secret.metadata.name]} />
                  <span>{secret.metadata.name}</span>
                </label>
              </div>
            );
          })
        ) : (
          <span className="margin-b-small">No existing credentials found.</span>
        )}
      </div>
    );
  }
}

export default AppRepoAddDockerCreds;
