import * as React from "react";
import { NavLink } from "react-router-dom";
import logo from "../../logo.svg";

import { INamespaceState } from "../../reducers/namespace";
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
  pathname: string;
  push: (path: string) => void;
  setNamespace: (ns: string) => void;
  hideLogoutLink: boolean;
}

interface IHeaderState {
  configOpen: boolean;
  mobileOpen: boolean;
}

class Header extends React.Component<IHeaderProps, IHeaderState> {
  // Defines the route
  private readonly links: IHeaderLinkProps[] = [
    {
      children: "Applications",
      exact: true,
      namespaced: true,
      to: "/apps",
    },
    {
      children: "Catalog",
      to: "/catalog",
    },
    {
      children: "Service Instances (alpha)",
      namespaced: true,
      to: "/services/instances",
    },
  ];

  constructor(props: any) {
    super(props);

    this.state = {
      configOpen: false,
      mobileOpen: false,
    };
  }

  public componentWillReceiveProps(newProps: IHeaderProps) {
    if (newProps.pathname !== this.props.pathname) {
      this.setState({ configOpen: false, mobileOpen: false });
    }
  }

  public render() {
    const { fetchNamespaces, namespace, authenticated: showNav, hideLogoutLink } = this.props;
    const header = `header ${this.state.mobileOpen ? "header-open" : ""}`;
    const submenu = `header__nav__submenu ${
      this.state.configOpen ? "header__nav__submenu-open" : ""
    }`;

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
                  onChange={this.handleNamespaceChange}
                  fetchNamespaces={fetchNamespaces}
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
                        <NavLink to="/config/repos">App Repositories</NavLink>
                      </li>
                      <li role="none">
                        <NavLink to="/config/brokers">Service Brokers</NavLink>
                      </li>
                    </ul>
                  </li>
                  {!hideLogoutLink && (
                    <li>
                      <NavLink to="#" onClick={this.handleLogout}>
                        <LogOut size={16} className="icon margin-r-tiny" /> Logout
                      </NavLink>
                    </li>
                  )}
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
    const { pathname, push, setNamespace } = this.props;
    const to = pathname.replace(/\/ns\/[^/]*/, `/ns/${ns}`);
    if (to === pathname) {
      setNamespace(ns);
    } else {
      push(to);
    }
  };
}

export default Header;
