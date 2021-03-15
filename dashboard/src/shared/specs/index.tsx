/* eslint-disable @typescript-eslint/no-var-requires */
export default (sharedExampleName: string, args: any) => {
  const sharedExamples = require(`./${sharedExampleName}`);
  sharedExamples.default(args);
};
