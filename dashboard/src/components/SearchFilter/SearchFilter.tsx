// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import Column from "components/js/Column";
import Row from "components/js/Row";
import React from "react";
import Input from "../js/Input";
import "./SearchFilter.css";

export interface ISearchFilterProps {
  value: string;
  className?: string;
  placeholder: string;
  onChange: (filter: string) => void;
  submitFilters: () => void;
}

function SearchFilter(props: ISearchFilterProps) {
  const [value, setValue] = React.useState(props.value);
  const [timeout, setTimeoutState] = React.useState({} as NodeJS.Timeout);
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    // Copy is needed to avoid losing the reference
    const valueCopy = e.currentTarget.value;
    setValue(e.currentTarget.value);
    // Gather changes before submitting
    clearTimeout(timeout);
    setTimeoutState(
      setTimeout(() => {
        props.onChange(valueCopy);
      }, 300),
    );
  };
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    props.submitFilters();
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
              value={value}
              {...Input.defaultProps}
            />
          </Column>
        </Row>
      </form>
    </div>
  );
}

export default SearchFilter;
