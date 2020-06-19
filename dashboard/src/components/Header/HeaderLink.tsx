import * as React from "react";
import { NavLink } from "react-router-dom";

// Expected props
export interface IHeaderLinkProps {
  to: string;
  exact?: boolean;
  children?: React.ReactChildren | React.ReactNode | string;
}

class HeaderLink extends React.Component<IHeaderLinkProps> {
  public static defaultProps: Partial<IHeaderLinkProps> = {
    exact: false,
  };

  public render() {
    const { to } = this.props;

    return (
      <NavLink
        to={to}
        activeClassName="header__nav__menu__item-active"
        className="header__nav__menu__item"
        exact={this.props.exact}
      >
        {this.props.children}
      </NavLink>
    );
  }
}

export default HeaderLink;
