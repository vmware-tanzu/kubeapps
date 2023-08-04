// Copyright 2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useRef } from "react";
import { ActionType, deprecated } from "typesafe-actions";
import { useNavigate, useLocation, Location } from "react-router-dom";
import { useDispatch } from "react-redux";

// When upgrading react-router from v5 to v6 (see #6187) we also needed to remove
// connected-react-router. One of the main uses was being able to dispatch a `push`
// action so that different reducers could watch for a change in URL state. This
// module exists to ease the transition, using the new `navigate` API from
// react-router v6 together with a custom LOCATION_CHANGE action that is only
// triggered if the location.pathname changes..

const { createAction } = deprecated;

export const LOCATION_CHANGE = "LOCATION_CHANGE";

export const locationChange = createAction(LOCATION_CHANGE, resolve => {
  return (resource: { location: Location }) => resolve(resource);
});

export type PushAction = ActionType<typeof locationChange>;

export function usePush() {
  const navigate = useNavigate();
  const location = useLocation();
  const dispatch = useDispatch();

  // We avoid triggering a location change on the first render of the hook.
  const firstRun = useRef(true);

  useEffect(() => {
    if (!firstRun.current) {
      dispatch(locationChange({ location }));
    } else {
      firstRun.current = false;
    }
  }, [dispatch, location]);

  return (url: string) => {
    navigate(url);
  };
}
