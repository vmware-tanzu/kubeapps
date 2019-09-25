import * as React from "react";
import Tab from "./Tab";

export interface ITab {
  title: string;
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
                title={tab.title}
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
