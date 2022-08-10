// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import MultiCheckbox from "components/MultiCheckbox";
import React from "react";

interface IFilterGroupProps {
  name: string;
  options: string[];
  currentFilters: string[];
  onAddFilter: (type: string, value: string) => void;
  onRemoveFilter: (etype: string, value: string) => void;
  disabled?: boolean;
}

function FilterGroup({
  name,
  currentFilters,
  options,
  disabled = false,
  onAddFilter,
  onRemoveFilter,
}: IFilterGroupProps) {
  const onChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { value: targetValue } = e.target;
    const index = currentFilters.indexOf(targetValue);
    if (index > -1) {
      // Already checked, remove
      onRemoveFilter(name, targetValue);
    } else {
      // Not checked, add
      onAddFilter(name, targetValue);
    }
  };
  return (
    <MultiCheckbox
      name={name}
      options={options}
      value={currentFilters}
      onChange={onChange}
      disabled={disabled}
    />
  );
}

export default FilterGroup;
