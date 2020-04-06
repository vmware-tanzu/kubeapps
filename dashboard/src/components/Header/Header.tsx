import * as React from "react";
import { NavLink } from "react-router-dom";
import logo from "../../logo.svg";

import { INamespaceState } from "../../reducers/namespace";
import { definedNamespaces } from "../../shared/Namespace";
import HeaderLink, { IHeaderLinkProps } from "./HeaderLink";
import NamespaceSelector from "./NamespaceSelector";

// Icons
import { LogOut, Settings } from "react-feather";

import "react-select/dist/react-select.css";
import "./Header.css";

interface IHeaderProps {
  authenticated: boolean;
  fetchNamespaces: () => void;
  logout: () => void;
  namespace: INamespaceState;
  defaultNamespace: string;
  pathname: string;
  push: (path: string) => void;
  setNamespace: (ns: string) => void;
  createNamespace: (ns: string) => Promise<boolean>;
  getNamespace: (ns: string) => void;
  featureFlags: { reposPerNamespace: boolean; operators: boolean };
}

interface IHeaderState {
  configOpen: boolean;
  mobileOpen: boolean;
}

class Header extends React.Component<IHeaderProps, IHeaderState> {
  public static defaultProps = {
    featureFlags: { reposPerNamespace: false, operators: false },
  };

  // Defines the route
  private readonly links: IHeaderLinkProps[] = [
    {
      children: "Applications",
      exact: true,
      namespaced: true,
      to: "apps",
    },
    {
      children: "Catalog",
      namespaced: true,
      to: "catalog",
    },
    {
      children: "Service Instances (alpha)",
      namespaced: true,
      to: "services/instances",
    },
  ];

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
      namespace,
      defaultNamespace,
      authenticated: showNav,
      createNamespace,
      getNamespace,
    } = this.props;
    const header = `header ${this.state.mobileOpen ? "header-open" : ""}`;
    const submenu = `header__nav__submenu ${
      this.state.configOpen ? "header__nav__submenu-open" : ""
    }`;

    const reposPath = this.props.featureFlags.reposPerNamespace
      ? `/config/ns/${namespace.current}/repos`
      : "/config/repos";
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
                  {this.links.map(link => (
                    <li key={link.to}>
                      <HeaderLink {...link} currentNamespace={namespace.current} />
                    </li>
                  ))}
                </ul>
              </nav>
            )}
            {showNav && (
              <div className="header__nav header__nav-config">
                <NamespaceSelector
                  namespace={namespace}
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
                          <NavLink to={`/ns/${namespace.current}/operators`}>Operators</NavLink>
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
