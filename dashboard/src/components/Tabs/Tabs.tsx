import * as React from "react";
import Tab from "./Tab";

import "./Tabs.css";

export interface ITab {
  header: string;
  content: JSX.Element;
}

export interface ITabsProps {
  tabs: ITab[];
}

export interface ITabsState {
  tabActive: number;
}

class Tabs extends React.Component<ITabsProps, ITabsState> {
  public state: ITabsState = {
    tabActive: 0,
  };

  public render() {
    const { tabs } = this.props;
    const { tabActive } = this.state;
    return (
      <>
        <div className="Tabs" role="tablist">
          {tabs.map((tab, i) => {
            return (
              <Tab
                key={`tab-${i}`}
                header={tab.header}
                active={i === tabActive}
                onClick={this.activeTab(i)}
              />
            );
          })}
        </div>
        <div>
          {tabs.map((tab, i) => {
            const active = i === tabActive;
            const id = `tab-${i}-content`;
            return (
              <div key={id} id={id} role="tabpanel" hidden={!active} aria-expanded={active}>
                {tab.content}
              </div>
            );
          })}
        </div>
      </>
    );
  }

  private activeTab = (index: number) => {
    return () => {
      this.setState({ tabActive: index });
    };
  };
}

export default Tabs;
