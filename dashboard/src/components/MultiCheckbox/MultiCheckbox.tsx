// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsCheckbox } from "@cds/react/checkbox/";
import "./MultiCheckbox.scss";

export interface IMultiCheckboxProps {
  name: string;
  value: string[];
  options: string[];
  disabled?: boolean;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

export function MultiCheckbox({ name, value, options, disabled, onChange }: IMultiCheckboxProps) {
  return (
    <div className="multicheckbox-wrapper">
      {options.map((opt, i) => (
        <div key={i}>
          <CdsCheckbox>
            <label htmlFor={`${name}-${i}`}>{opt}</label>
            <input
              type="checkbox"
              value={opt}
              id={`${name}-${i}`}
              checked={value.includes(opt)}
              disabled={disabled}
              onChange={onChange}
            />
          </CdsCheckbox>
        </div>
      ))}
    </div>
  );
}
