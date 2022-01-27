// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Switch from "react-switch";
import { IBasicFormParam } from "shared/types";

export interface IStringParamProps {
  id: string;
  label: string;
  param: IBasicFormParam;
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

function BooleanParam({ id, param, label, handleBasicFormParamChange }: IStringParamProps) {
  // handleChange transform the event received by the Switch component to a checkbox event
  const handleChange = (checked: boolean) => {
    const event = {
      currentTarget: { value: String(checked), type: "checkbox", checked },
    } as React.FormEvent<HTMLInputElement>;
    handleBasicFormParamChange(param)(event);
  };

  return (
    <label htmlFor={id}>
      <div>
        <Switch
          height={20}
          width={40}
          id={id}
          onChange={handleChange}
          checked={!!param.value}
          className="react-switch"
          onColor="#5aa220"
          checkedIcon={false}
          uncheckedIcon={false}
        />
        <span className="deployment-form-label">{label}</span>
      </div>
      {param.description && <span className="description">{param.description}</span>}
    </label>
  );
}

export default BooleanParam;
