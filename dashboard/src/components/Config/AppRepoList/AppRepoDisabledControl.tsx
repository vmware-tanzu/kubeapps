// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import "./AppRepoControl.css";

export function AppRepoDisabledControl() {
  return (
    <div className="apprepo-control-buttons">
      <CdsButton disabled={true} action="outline">
        Edit
      </CdsButton>
      <CdsButton disabled={true} action="outline">
        Refresh
      </CdsButton>
      <CdsButton disabled={true} action="outline">
        Delete
      </CdsButton>
    </div>
  );
}
