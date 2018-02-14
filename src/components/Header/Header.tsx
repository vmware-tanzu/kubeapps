import * as React from "react";
import { NavLink } from "react-router-dom";
import logo from "../../logo.svg";

import HeaderLink, { IHeaderLinkProps } from "./HeaderLink";

// Icons
import Cog from "!react-svg-loader!open-iconic/svg/cog.svg";

import "./Header.css";

class Header extends React.Component {
  // Defines the route
  private readonly links: IHeaderLinkProps[] = [
    {
      children: "Applications",
      exact: true,
      to: "/",
    },
    {
      children: "Charts",
      to: "/charts",
    },
    {
      children: "Service Catalog",
      to: "/services",
    },
    {
      children: "App Repositories",
      to: "/repos",
    },
  ];

  public render() {
    return (
      <section className="gradient-135-brand type-color-reverse type-color-reverse-anchor-reset">
        <div className="container">
          <header className="header">
            <div className="header__logo">
              <NavLink to="/">
                <img src={logo} alt="Kubeapps logo" />
              </NavLink>
            </div>
            <nav className="header__nav">
              <ul className="header__nav__menu" role="menubar">
                {this.links.map(link => (
                  <li key={link.to}>
                    <HeaderLink {...link} />
                  </li>
                ))}
              </ul>
            </nav>
            <div className="header__nav__user">
              <HeaderLink to="/configuration">
                <Cog className="icon icon-small margin-r-tiny" /> Configuration
              </HeaderLink>
            </div>
          </header>
        </div>
      </section>
    );
  }
}

export default Header;
