import { useEffect, useState } from "react";
import placeholder from "../../placeholder.png";

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
