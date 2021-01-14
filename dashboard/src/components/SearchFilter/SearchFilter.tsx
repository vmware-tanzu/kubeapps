import { CdsIcon } from "@clr/react/icon";
import * as React from "react";
import Input from "../js/Input";

import Column from "components/js/Column";
import Row from "components/js/Row";
import "./SearchFilter.css";

export interface ISearchFilterProps {
  value: string;
  className?: string;
  placeholder: string;
  onChange: (filter: string) => void;
  submitFilters: (filter: string) => void;
}

function SearchFilter(props: ISearchFilterProps) {
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    props.onChange(e.currentTarget.value);
  };

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    props.submitFilters((e.currentTarget.elements[0] as HTMLInputElement).value);
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
