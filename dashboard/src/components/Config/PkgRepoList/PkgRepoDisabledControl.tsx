// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import "./PkgRepoControl.css";

export function PkgRepoDisabledControl() {
  return (
    <div className="pkgrepo-control-buttons">
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
