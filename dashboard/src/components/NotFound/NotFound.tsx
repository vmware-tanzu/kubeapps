// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import logo404 from "../../img/404.svg";

function NotFound() {
  return (
    <div className="section-not-found">
      <div>
        <img src={logo404} alt="Kubeapps logo" />
        <h3>The page you are looking for can't be found.</h3>
      </div>
    </div>
  );
}

export default NotFound;
