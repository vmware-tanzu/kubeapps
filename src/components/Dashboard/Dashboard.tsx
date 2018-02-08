import * as React from "react";
import { Link } from "react-router-dom";

class Dashboard extends React.Component {
  public render() {
    return (
      <section className="Dashboard">
        <header className="Dashboard__header">
          <h1>Apps</h1>
          <hr />
        </header>
        <main className="text-c">
          <div>No Apps installed</div>
          <div className="padding-normal">
            <Link className="button button-primary" to="/charts">
              Deploy Chart
            </Link>
          </div>
        </main>
      </section>
    );
  }
}

export default Dashboard;
