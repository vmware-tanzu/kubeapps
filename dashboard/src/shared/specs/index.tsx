// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

/* eslint-disable @typescript-eslint/no-var-requires */
export default (sharedExampleName: string, args: any) => {
  const sharedExamples = require(`./${sharedExampleName}`);
  sharedExamples.default(args);
};
