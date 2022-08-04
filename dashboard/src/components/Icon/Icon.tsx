// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from "react";
import placeholder from "icons/placeholder.svg";

export interface IIconProps {
  icon?: any;
}

function Icon({ icon }: IIconProps) {
  const [srcIcon, setSrcIcon] = useState(placeholder);
  const [iconErrored, setIconErrored] = useState(false);
  useEffect(() => {
    if (srcIcon !== icon && icon && !iconErrored) {
      setSrcIcon(icon);
    }
  }, [srcIcon, icon, iconErrored]);

  const onError = () => {
    setIconErrored(true);
    setSrcIcon(placeholder);
  };

  return <img src={srcIcon} alt="icon" onError={onError} />;
}

export default Icon;
