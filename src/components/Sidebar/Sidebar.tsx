import * as React from "react";
import { Link } from "react-router-dom";

import placeholder from "../../placeholder.png";
import "./Sidebar.css";

const sidebarItem = (props: { to: string; text: string; imageUrl?: string }) => {
  const { to, text, imageUrl } = props;
  const imageSrc: string = imageUrl || placeholder;

  return (
    <li className="padding-v-normal">
      <Link to={to}>
        <img src={imageSrc} height="48" />
        <div className="type-small">
          <span>{text}</span>
        </div>
      </Link>
    </li>
  );
};

class Sidebar extends React.Component {
  public render() {
    return (
      <aside className="Sidebar bg-dark type-color-reverse-anchor-reset">
        <ul className="remove-style margin-reset padding-h-normal text-c">
          <li className="padding-v-normal">
            <Link to="/">
              <img src={placeholder} height="48" />
              <div className="type-small">Apps</div>
            </Link>
          </li>
          <li className="padding-v-normal">
            <img src={placeholder} height="48" />
            <div className="type-small">Functions</div>
          </li>
          {sidebarItem({ to: "/charts", text: "Charts" })}
          {sidebarItem({ to: "/services", text: "Service Catalog" })}
          {sidebarItem({ to: "/repos", text: "App Repositories" })}
        </ul>
      </aside>
    );
  }
}

export default Sidebar;
