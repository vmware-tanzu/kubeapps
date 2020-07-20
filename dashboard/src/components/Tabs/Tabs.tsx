import React, { useState } from "react";

import "./Tabs.scss";

interface ITabsProps {
  columns: string[];
  data: Array<string | JSX.Element | JSX.Element[]>;
}

export default function Tabs({ columns, data }: ITabsProps) {
  const [selected, setSelected] = useState(0);
  const handleClick = (tab: number) => {
    return () => setSelected(tab);
  };
  return (
    <div className="tabs">
      <ul className="nav" role="tablist">
        {columns.map((column, index) => {
          return (
            <li role="presentation" className="nav-item" key={`${column}-${index}`}>
              <button
                id={`tab${index}`}
                className={`btn btn-link nav-link tab-button ${selected === index ? "active" : ""}`}
                aria-controls={`panel${index}`}
                type="button"
                onClick={handleClick(index)}
              >
                {column}
              </button>
            </li>
          );
        })}
      </ul>
      {data.map((children, index) => {
        return (
          <section
            key={`panel${index}`}
            id={`panel${index}`}
            role="tabpanel"
            aria-labelledby={`tab${index}`}
            aria-hidden={selected !== index ? "true" : "false"}
          >
            {children}
          </section>
        );
      })}
    </div>
  );
}
