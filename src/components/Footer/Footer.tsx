import * as React from "react";
import logo from "../../logo.svg";

class Footer extends React.Component {
  public render() {
    return (
      <footer className="osFooter bg-dark type-color-reverse-anchor-reset">
        <div className="container padding-h-big padding-v-bigger">
          <div className="row collapse-b-phone-land align-center">
            <div className="col-6">
              <h4 className="inverse margin-reset">
                <img src={logo} alt="Kubeapps logo" className="osFooter__logo" />
              </h4>
              <p className="type-color-white type-small margin-reset">
                Made with &#10084; by Bitnami and{" "}
                <a href="https://github.com/kubeapps/kubeapps/graphs/contributors" target="_blank">
                  contributors
                </a>
              </p>
            </div>
            <div className="col-6 text-r">
              <a href="#" className="socialIcon margin-h-small">
                <svg
                  role="img"
                  aria-label="See the Facebook Profile of Bitnami"
                  viewBox="0 0 54 54"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <title>facebook</title>
                  <circle fill="currentColor" cx="27" cy="27" r="27" />
                  <path
                    d="M23.723 40h5.235V26.89h3.653L33 22.5h-4.042V20c0-1.035.208-1.444 1.209-1.444H33V14h-3.625c-3.896 0-5.652 1.716-5.652 5v3.5H21v4.444h2.723V40z"
                    fill="currentColor"
                  />
                </svg>
              </a>
              <a href="#" className="socialIcon margin-h-small">
                <svg
                  role="img"
                  aria-label="See the Twitter Profile of Bitnami"
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
              <a href="#" className="socialIcon margin-h-small">
                <svg
                  role="img"
                  aria-label="See the Github Profile of Bitnami"
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
              <a href="#" className="socialIcon margin-h-small">
                <svg
                  role="img"
                  aria-label="See the Youtube Profile of Bitnami"
                  viewBox="0 0 54 54"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <title>youtube</title>
                  <circle fill="currentColor" cx="27" cy="27" r="27" />
                  <path
                    d="M24.2 31.286v-8.572L31.474 27 24.2 31.286zm16.215-11.163a3.543 3.543 0 0 0-2.476-2.526C35.756 17 27 17 27 17s-8.755 0-10.938.597a3.544 3.544 0 0 0-2.476 2.526C13 22.351 13 27 13 27s0 4.649.585 6.877a3.543 3.543 0 0 0 2.476 2.526C18.244 37 27 37 27 37s8.756 0 10.94-.597a3.543 3.543 0 0 0 2.475-2.526C41 31.649 41 27 41 27s0-4.649-.585-6.877z"
                    fill="currentColor"
                  />
                </svg>
              </a>
              <a href="#" className="socialIcon margin-h-small">
                <svg
                  role="img"
                  aria-label="See the LinkedIn Profile of Bitnami"
                  viewBox="0 0 54 54"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <title>linkedin</title>
                  <circle fill="currentColor" cx="27" cy="27" r="27" />
                  <path
                    d="M20.6 17.8c0 1.542-1.253 2.8-2.8 2.8S15 19.35 15 17.8c0-1.542 1.253-2.8 2.8-2.8s2.8 1.258 2.8 2.8zm0 5.2h-4.8v16h4.8V23zm7.889-.303H23.8V39h4.689v-8.553c0-2.295 1.024-3.656 2.979-3.656 1.802 0 2.666 1.309 2.666 3.656V39H39V28.676c0-4.364-2.395-6.476-5.755-6.476-3.351 0-4.765 2.697-4.765 2.697v-2.2h.009z"
                    fill="currentColor"
                  />
                </svg>
              </a>
            </div>
          </div>
        </div>
      </footer>
    );
  }
}

export default Footer;
