import * as React from "react";
import logo from "../../logo.svg";

// Icons
import Heart from "../../icons/Heart";

interface IFooterProps {
  appVersion: string;
}

const Footer: React.SFC<IFooterProps> = props => {
  return (
    <footer className="osFooter bg-dark type-color-reverse-anchor-reset">
      <div className="container padding-h-big padding-v-bigger">
        <div className="row collapse-b-phone-land align-center">
          <div className="col-6">
            <h4 className="inverse margin-reset">
              <img src={logo} alt="Kubeapps logo" className="osFooter__logo" />
            </h4>
            <p className="type-small margin-reset">
              Made with <Heart className="icon icon-small" /> by Bitnami and{" "}
              <a
                href="https://github.com/kubeapps/kubeapps/graphs/contributors"
                className="type-color-white"
                target="_blank"
              >
                contributors
              </a>
              .
            </p>
          </div>
          <div className="col-6 text-r">
            <a href="https://twitter.com/kubeapps" className="socialIcon margin-small">
              <svg
                role="img"
                aria-label="Kubeapps on Twitter"
                viewBox="0 0 54 54"
                xmlns="http://www.w3.org/2000/svg"
              >
                <title>twitter</title>
                <circle fill="currentColor" cx="27" cy="27" r="27" />
                <path
                  d="M14 35.618A15.166 15.166 0 0 0 22.177 38c9.904 0 15.498-8.313 15.162-15.77A10.761 10.761 0 0 0 40 19.485c-.957.422-1.985.707-3.063.834a5.314 5.314 0 0 0 2.344-2.932 10.729 10.729 0 0 1-3.386 1.287A5.344 5.344 0 0 0 32 17c-3.442 0-5.973 3.193-5.195 6.51a15.17 15.17 0 0 1-10.994-5.54 5.288 5.288 0 0 0 1.65 7.078 5.33 5.33 0 0 1-2.417-.663c-.057 2.456 1.714 4.753 4.279 5.265-.751.204-1.573.25-2.408.09a5.33 5.33 0 0 0 4.982 3.683A10.767 10.767 0 0 1 14 35.618"
                  fill="currentColor"
                />
              </svg>
            </a>
            <a href="https://github.com/kubeapps/kubeapps" className="socialIcon margin-small">
              <svg
                role="img"
                aria-label="Kubeapps on GitHub"
                viewBox="0 0 54 54"
                xmlns="http://www.w3.org/2000/svg"
              >
                <title>github</title>
                <circle fill="currentColor" cx="27" cy="27" r="27" />
                <path
                  d="M27.5 14C20.044 14 14 19.968 14 27.33c0 5.888 3.868 10.885 9.233 12.647.675.122.921-.289.921-.642 0-.317-.011-1.155-.018-2.268-3.755.806-4.547-1.786-4.547-1.786-.614-1.54-1.5-1.95-1.5-1.95-1.225-.827.094-.81.094-.81 1.355.094 2.067 1.373 2.067 1.373 1.204 2.038 3.16 1.45 3.93 1.108.122-.861.47-1.449.856-1.782-2.997-.336-6.149-1.48-6.149-6.588 0-1.455.526-2.644 1.39-3.576-.14-.337-.603-1.693.132-3.527 0 0 1.133-.36 3.712 1.366a13.085 13.085 0 0 1 3.38-.449c1.146.005 2.301.153 3.38.449 2.577-1.725 3.708-1.366 3.708-1.366.737 1.834.273 3.19.134 3.527.865.932 1.388 2.121 1.388 3.576 0 5.12-3.156 6.248-6.164 6.578.485.411.917 1.225.917 2.468 0 1.782-.017 3.22-.017 3.657 0 .356.243.77.928.64C37.135 38.21 41 33.218 41 27.33 41 19.968 34.955 14 27.5 14"
                  fill="currentColor"
                />
              </svg>
            </a>
            <p className="type-small margin-small version-link">
              <a
                href={`https://github.com/kubeapps/kubeapps/releases/tag/${props.appVersion}`}
                className="type-color-white"
                target="_blank"
              >
                {props.appVersion}
              </a>
            </p>
          </div>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
