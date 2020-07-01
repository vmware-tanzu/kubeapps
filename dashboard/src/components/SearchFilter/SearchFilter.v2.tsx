import { ClarityIcons, searchIcon } from "@clr/core/icon-shapes";
import * as React from "react";
import { CdsIcon } from "../Clarity/clarity";

import "./SearchFilter.v2.css";

ClarityIcons.addIcons(searchIcon);

export interface ISearchFilterProps {
  value: string;
  className?: string;
  placeholder: string;
  onChange: (filter: string) => void;
  onSubmit: (filter: string) => void;
}

function SearchFilter(props: ISearchFilterProps) {
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    props.onChange(e.currentTarget.value);
  };
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    props.onSubmit(props.value);
  };
  return (
    <div className="search-box">
      <form onSubmit={handleSubmit}>
        <CdsIcon size="sm" shape="search" />
        <input
          id="search"
          type="text"
          className="padding-l-bigger"
          placeholder={props.placeholder}
          autoComplete="off"
          onChange={handleChange}
          value={props.value}
        />
      </form>
    </div>
  );
}

export default SearchFilter;
