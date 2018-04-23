import * as React from "react";
import { NavLink } from "react-router-dom";

// Expected props
export interface IHeaderLinkProps {
  to: string;
  exact?: boolean;
  external?: boolean;
  children?: React.ReactChildren | React.ReactNode | string;
  currentNamespace?: string | null;
  namespaced?: boolean;
}

class HeaderLink extends React.Component<IHeaderLinkProps> {
  public static defaultProps: Partial<IHeaderLinkProps> = {
    exact: false,
    external: false,
  };

  public renderInternalLink() {
    const { currentNamespace, namespaced, to } = this.props;
    const link = currentNamespace && namespaced ? `${to}/ns/${currentNamespace}` : to;
    return (
      <NavLink
        to={link}
        activeClassName="header__nav__menu__item-active"
        className="header__nav__menu__item"
        exact={this.props.exact}
      >
        {this.props.children}
      </NavLink>
    );
  }

  public renderExternalLink() {
    return (
      <a href={this.props.to} className="header__nav__menu__item">
        {this.props.children}
      </a>
    );
  }

  public render() {
    return this.props.external ? this.renderExternalLink() : this.renderInternalLink();
  }
}

export default HeaderLink;
