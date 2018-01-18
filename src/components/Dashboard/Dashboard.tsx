import * as React from 'react';
import { Link } from 'react-router-dom';

class Dashboard extends React.Component {
  render() {
    return (
      <section className="Dashboard">
        <header className="Dashboard__header">
          <h1>Applications</h1>
          <hr />
        </header>
        <main className="text-c">
          <div>No Apps installed</div>
          <div className="padding-normal">
            <Link className="button button-primary" to="/charts">deploy one</Link>
          </div>
        </main>
      </section>
    );
  }
}

export default Dashboard;
