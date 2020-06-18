import * as React from "react";
import { LogOut, Settings } from "react-feather";
import { NavLink } from "react-router-dom";
import "react-select/dist/react-select.css";
import logo from "../../logo.svg";
import { IClusterState } from "../../reducers/namespace";
import { definedNamespaces } from "../../shared/Namespace";
import { app } from "../../shared/url";
import "./Header.css";
import HeaderLink from "./HeaderLink";
import NamespaceSelector from "./NamespaceSelector";

interface IHeaderProps {
  authenticated: boolean;
  fetchNamespaces: () => void;
  logout: () => void;
  cluster: IClusterState;
  defaultNamespace: string;
  pathname: string;
  push: (path: string) => void;
  setNamespace: (ns: string) => void;
  createNamespace: (ns: string) => Promise<boolean>;
  getNamespace: (ns: string) => void;
  featureFlags: { operators: boolean };
}

interface IHeaderState {
  configOpen: boolean;
  mobileOpen: boolean;
}

class Header extends React.Component<IHeaderProps, IHeaderState> {
  public static defaultProps = {
    featureFlags: { operators: false },
  };

  constructor(props: any) {
    super(props);

    this.state = {
      configOpen: false,
      mobileOpen: false,
    };
  }

  public componentDidUpdate(prevProps: IHeaderProps) {
    if (prevProps.pathname !== this.props.pathname) {
      this.setState({
        configOpen: false,
        mobileOpen: false,
      });
    }
  }

  public render() {
    const {
      fetchNamespaces,
      cluster: namespace,
      defaultNamespace,
      authenticated: showNav,
      createNamespace,
      getNamespace,
    } = this.props;
    const header = `header ${this.state.mobileOpen ? "header-open" : ""}`;
    const submenu = `header__nav__submenu ${
      this.state.configOpen ? "header__nav__submenu-open" : ""
    }`;

    const reposPath = `/config/ns/${namespace.currentNamespace}/repos`;
    return (
      <section className="gradient-135-brand type-color-reverse type-color-reverse-anchor-reset">
        <div className="container">
          <header className={header}>
            <div className="header__logo">
              <NavLink to="/">
                <img src={logo} alt="Kubeapps logo" />
              </NavLink>
            </div>
            {showNav && (
              <nav className="header__nav">
                <button
                  className="header__nav__hamburguer"
                  aria-label="Menu"
                  aria-haspopup="true"
                  aria-expanded="false"
                  onClick={this.toggleMobile}
                >
                  <div />
                  <div />
                  <div />
                </button>
                <ul className="header__nav__menu" role="menubar">
                  <li>
                    <HeaderLink to={app.apps.list(namespace.currentNamespace)} exact={true}>Applications</HeaderLink>
                  </li>
                  <li>
                    <HeaderLink to={app.catalog(namespace.currentNamespace)}>Catalog</HeaderLink>
                  </li>
                  <li>
                    <HeaderLink to={app.servicesInstances(namespace.currentNamespace)}>Service Instances (alpha)</HeaderLink>
                  </li>
                </ul>
              </nav>
            )}
            {showNav && (
              <div className="header__nav header__nav-config">
                <NamespaceSelector
                  cluster={namespace}
                  defaultNamespace={defaultNamespace}
                  onChange={this.handleNamespaceChange}
                  fetchNamespaces={fetchNamespaces}
                  createNamespace={createNamespace}
                  getNamespace={getNamespace}
                />
                <ul className="header__nav__menu" role="menubar">
                  <li
                    onMouseEnter={this.openSubmenu}
                    onMouseLeave={this.closeSubmenu}
                    onClick={this.toggleSubmenu}
                  >
                    <a>
                      <Settings size={16} className="icon margin-r-tiny" /> Configuration
                    </a>
                    <ul role="menu" aria-label="Products" className={submenu}>
                      <li role="none">
                        <NavLink to={reposPath}>App Repositories</NavLink>
                      </li>
                      <li role="none">
                        <NavLink to="/config/brokers">Service Brokers</NavLink>
                      </li>
                      {this.props.featureFlags.operators && (
                        <li role="none">
                          <NavLink to={`/ns/${namespace.currentNamespace}/operators`}>Operators</NavLink>
                        </li>
                      )}
                    </ul>
                  </li>
                  <li>
                    <NavLink to="#" onClick={this.handleLogout}>
                      <LogOut size={16} className="icon margin-r-tiny" /> Logout
                    </NavLink>
                  </li>
                </ul>
              </div>
            )}
          </header>
        </div>
      </section>
    );
  }

  private isTouchDevice = (): boolean => {
    return "ontouchstart" in document.documentElement;
  };

  private toggleMobile = () => {
    this.setState({ mobileOpen: !this.state.mobileOpen });
  };

  private openSubmenu = () => {
    if (!this.isTouchDevice()) {
      this.setState({ configOpen: true });
    }
  };

  private closeSubmenu = () => {
    if (!this.isTouchDevice()) {
      this.setState({ configOpen: false });
    }
  };

  private toggleSubmenu = () => {
    this.setState({ configOpen: !this.state.configOpen });
  };

  private handleLogout = (e: React.MouseEvent<HTMLAnchorElement>) => {
    e.preventDefault();
    this.props.logout();
  };

  private handleNamespaceChange = (ns: string) => {
    const { pathname, push, setNamespace, getNamespace } = this.props;
    const to = pathname.replace(/\/ns\/[^/]*/, `/ns/${ns}`);
    setNamespace(ns);
    if (ns !== definedNamespaces.all) {
      getNamespace(ns);
    }
    if (to !== pathname) {
      push(to);
    }
  };
}

export default Header;
