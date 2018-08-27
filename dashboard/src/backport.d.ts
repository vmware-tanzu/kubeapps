// This file contains a set of backports from typescript 3.x
// TODO(miguel) Remove backports once we upgrade typescript https://github.com/kubeapps/kubeapps/issues/534
type EventHandlerNonNull = (event: Event) => any;

// @ts-ignore
type BinaryType = "blob" | "arraybuffer";
