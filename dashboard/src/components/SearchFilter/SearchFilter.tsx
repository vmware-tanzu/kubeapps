import * as React from "react";
import { Search } from "react-feather";

import "./SearchFilter.css";

interface ISearchFilterProps {
  value: string;
  className?: string;
  placeholder: string;
  onChange: (filter: string) => void;
  onSubmit: (filter: string) => void;
}

class SearchFilter extends React.Component<ISearchFilterProps> {
  public render() {
    return (
      <div className={`SearchFilter ${this.props.className}`}>
        <Search size={16} />
        <form onSubmit={this.handleSubmit}>
          <input
            id="search"
            type="text"
            className="padding-l-bigger"
            placeholder={this.props.placeholder}
            autoComplete="off"
            onChange={this.handleChange}
            value={this.props.value}
          />
        </form>
      </div>
    );
  }

  private handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.props.onChange(e.currentTarget.value);
  };
  private handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    this.props.onSubmit(this.props.value);
  };
}

export default SearchFilter;
