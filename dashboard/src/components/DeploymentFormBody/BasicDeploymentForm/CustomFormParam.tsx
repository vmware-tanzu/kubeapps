import React, { useMemo } from "react";
import { IBasicFormParam } from "shared/types";
import { CustomComponent } from "../../../RemoteComponent";

export interface ICustomParamProps {
  param: IBasicFormParam;
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

export default function CustomFormComponentLoader({
  param,
  handleBasicFormParamChange,
}: ICustomParamProps) {
  // Fetches the custom-component bundle served by the dashboard nginx
  const url = `${window.location.origin}/custom_components.js`;

  return useMemo(
    () => (
      <CustomComponent
        url={url}
        param={param}
        handleBasicFormParamChange={handleBasicFormParamChange}
      />
    ),
    [handleBasicFormParamChange, param, url],
  );
}
