import * as React from "react";
import { CdsIcon } from "../Clarity/clarity";
import Input from "../js/Input";

import Column from "components/js/Column";
import Row from "components/js/Row";
import "./SearchFilter.v2.css";

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
      <form onSubmit={handleSubmit} role="search" aria-label="Search on the site">
        <Row>
          <Column span={1}>
            <CdsIcon size="sm" shape="search" />
          </Column>
          <Column span={10}>
            <Input
              id="search"
              name="search"
              type="text"
              placeholder={props.placeholder}
              autoComplete="off"
              onChange={handleChange}
              value={props.value}
              {...Input.defaultProps}
            />
          </Column>
        </Row>
      </form>
    </div>
  );
}

export default SearchFilter;
