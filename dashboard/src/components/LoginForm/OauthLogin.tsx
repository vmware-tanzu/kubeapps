import { CdsButton } from "@clr/react/button";

interface ILoginFormProps {
  authenticationError: string | undefined;
  oauthLoginURI: string;
}

function OAuthLogin(props: ILoginFormProps) {
  return (
    <section className="title" aria-labelledby="login-title" aria-describedby="login-desc">
      <h3 id="login-title" className="welcome">
        Welcome to <span>Kubeapps</span>
      </h3>
      <p id="login-desc" className="hint">
        Your cluster operator has enabled login via an authentication provider.
      </p>
      <div className="login-group">
        {props.authenticationError && (
          <div className="error active">
            There was an error connecting to the Kubernetes API. Please check that your token is
            valid.
          </div>
        )}
        <a href={props.oauthLoginURI} className="login-submit-button">
          <CdsButton id="login-submit-button" status="primary">
            Login via OIDC Provider
          </CdsButton>
        </a>
      </div>
    </section>
  );
}

export default OAuthLogin;
