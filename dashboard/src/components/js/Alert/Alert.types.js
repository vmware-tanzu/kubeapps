// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { AlertThemes } from "./Alert";

/**
 * Types for alerts
 */

const customIcon = (props, propName, componentName) => {
  const value = props[propName];
  const { app } = props;

  if (app !== true && typeof value === "string" && value.length > 0) {
    return new Error(
      `Invalid ${propName} for ${componentName}. Custom Icon is only available for app level alerts`,
    );
  } else if (value != null && typeof value !== "string") {
    return new Error(
      `Invalid ${propName} for ${componentName}. Custom Icon property must be an String.`,
    );
  } else {
    return null;
  }
};

const theme = (props, propName, componentName) => {
  const value = props[propName];
  const { app } = props;
  const expectedValues = app
    ? Object.values(AlertThemes).filter(v => v !== "success")
    : Object.values(AlertThemes);

  if (expectedValues.includes(value)) {
    return null;
  } else {
    let error = `Invalid ${propName} for ${componentName}. Available values are: ${expectedValues.join(
      ", ",
    )}`;
    if (app && value === "success") {
      error += ". Remember success is not a valid value for App level Alerts.";
    }
    return new Error(error);
  }
};

export default {
  theme,
  customIcon,
};
