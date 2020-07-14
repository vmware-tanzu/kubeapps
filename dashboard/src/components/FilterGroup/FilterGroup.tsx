import React, { useState } from "react";

import MultiCheckbox from "components/js/MultiCheckbox";

interface IFilterGroupProps {
  name: string;
  options: string[];
  onChange: (newValue: string[]) => void;
}

function FilterGroup({ name, options, onChange: propsOnChange }: IFilterGroupProps) {
  const [value, setValue] = useState([] as string[]);
  const onChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { value: targetValue } = e.target;
    const index = value.indexOf(targetValue);
    let newValue = [];
    if (index > -1) {
      // Already checked, remove
      value.splice(index, 1);
      newValue = [...value];
    } else {
      // Not checked, add
      newValue = value.concat(targetValue);
    }
    setValue(newValue);
    propsOnChange(newValue);
  };
  return <MultiCheckbox name={name} options={options} span={1} value={value} onChange={onChange} />;
}

export default FilterGroup;
