import MultiCheckbox from "components/js/MultiCheckbox";
import React from "react";

interface IFilterGroupProps {
  name: string;
  options: string[];
  currentFilters: string[];
  onAddFilter: (type: string, value: string) => void;
  onRemoveFilter: (etype: string, value: string) => void;
}

function FilterGroup({
  name,
  currentFilters,
  options,
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
      span={1}
      value={currentFilters}
      onChange={onChange}
    />
  );
}

export default FilterGroup;
