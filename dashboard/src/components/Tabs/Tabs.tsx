// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { useState } from "react";
import "./Tabs.css";

interface ITabsProps {
  id: string;
  // Columns contains an array of tuples, where the second tuple element is
  // a function which should be called when the tab is activated. This is
  // necessary because after an update, the onclick events of passed tab columns
  // are not being fired.
  columns: Array<[string | JSX.Element, () => void]>;
  data: Array<string | JSX.Element | JSX.Element[]>;
}

export default function Tabs({ id, columns, data }: ITabsProps) {
  const [selected, setSelected] = useState(0);
  const handleClick = (tab: number, eventHandler: () => void) => {
    return () => {
      eventHandler();
      setSelected(tab);
    };
  };
  return (
    <div className="tabs">
      <ul className="nav" role="tablist" id={id}>
        {columns.map(([column, eventHandler], index) => {
          return (
            <li role="presentation" className="nav-item" key={`${column}-${index}`}>
              <button
                id={`${id}-tab${index}`}
                className={`btn btn-link nav-link tab-button ${selected === index ? "active" : ""}`}
                aria-controls={`${id}-panel${index}`}
                type="button"
                onClick={handleClick(index, eventHandler)}
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
            key={`${id}-panel${index}`}
            id={`${id}-panel${index}`}
            role="tabpanel"
            aria-labelledby={`${id}-tab${index}`}
            aria-hidden={selected !== index ? "true" : "false"}
          >
            {children}
          </section>
        );
      })}
    </div>
  );
}
