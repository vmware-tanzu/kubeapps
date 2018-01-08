import * as React from 'react';
import { Link } from 'react-router-dom';

const logo = require('../logo.svg');

class Header extends React.Component {
  render() {
    return (
      <section className="gradient-135-brand elevation-1 type-color-reverse-anchor-reset">
        <header className="OSHeader padding-v-normal padding-h-small">
          <div className="OSHeader__Logo">
            <Link to="/">
              <img src={logo} alt="Logo white" />
            </Link>
          </div>
        </header>
      </section>
    );
  }
}

export default Header;
