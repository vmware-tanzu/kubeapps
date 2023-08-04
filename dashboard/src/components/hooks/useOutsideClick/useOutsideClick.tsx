// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { Dispatch, MutableRefObject, SetStateAction, useCallback, useEffect } from "react";

/**
 * Detects when there's a click event outside the given element
 *
 * @param setIsOpen function to dispatch when users click outside the element
 * @param refs list of ref React objects that references an element in the DOM
 * @param enabled controls when the even listener should be added or not
 */
const useOutsideClick = (
  setIsOpen: Dispatch<SetStateAction<boolean>>,
  refs: MutableRefObject<HTMLElement | null>[],
  enabled = true,
) => {
  const memoizeClick = useCallback(
    e => {
      const clickedOutside =
        refs &&
        refs.every(ref => {
          return ref.current && !ref.current.contains(e.target);
        });

      if (clickedOutside) {
        setIsOpen(false);
      }
    },
    [setIsOpen, refs],
  );

  useEffect(() => {
    if (enabled) {
      document.addEventListener("mousedown", memoizeClick, { capture: true });
      document.addEventListener("touchstart", memoizeClick, { capture: true });
    }
    return () => {
      document.removeEventListener("mousedown", memoizeClick);
      document.removeEventListener("touchstart", memoizeClick);
    };
  }, [memoizeClick, enabled]);
};

export default useOutsideClick;
