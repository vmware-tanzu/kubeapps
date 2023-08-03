// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { HTMLAttributes } from "react";

export interface ICardProps {
  children: React.ReactNode;
  clickable?: boolean;
  onClick?: (e: React.MouseEvent<HTMLDivElement>) => void;
}

const Card = ({ children, clickable = false, onClick }: ICardProps) => {
  // Common props for the element
  const innerProps: HTMLAttributes<HTMLDivElement> = {
    className: `card ${clickable || typeof onClick === "function" ? "clickable" : ""}`,
    onClick: onClick,
  };

  // Make the card focusable by keyboard navigation. I'm not adding it based on the
  // clickable prop because I assume there's something around the card that manages
  // the onClick event.
  if (typeof onClick === "function") {
    innerProps.tabIndex = 0;

    // Runs onClick when the user types `enter` or `space`
    innerProps.onKeyDown = e => {
      // Enter
      if (e.key === "Enter" || e.key === " ") {
        onClick(e as any);
      }
    };
  }

  return <div {...innerProps}>{children}</div>;
};

export default Card;
