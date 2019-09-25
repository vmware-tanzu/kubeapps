import * as React from "react";

export interface ITabProps {
  key: string;
  header: string;
  active: boolean;
  onClick: () => void;
}

class Tab extends React.Component<ITabProps> {
  public render() {
    const { header, active, onClick, key } = this.props;
    return (
      <div className={`Tabs__Tab ${active ? "Tabs__Tab-active" : ""}`}>
        <button
          type="button"
          onClick={onClick}
          role="tab"
          aria-controls={`${key}-content`}
          aria-selected={active}
        >
          {header}
        </button>
      </div>
    );
  }
}

export default Tab;
