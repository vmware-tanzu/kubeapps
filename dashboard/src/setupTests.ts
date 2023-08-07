/* eslint-disable no-console */
// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Adapter from "@wojtekmaj/enzyme-adapter-react-17";
import Enzyme from "enzyme";
import "jest-enzyme";
import { WebSocket } from "mock-socket";
import ResizeObserver from "resize-observer-polyfill";
import { TextDecoder, TextEncoder } from "util";

Enzyme.configure({ adapter: new Adapter() });

// Mock browser specific APIs like localstorage or Websocket
jest.spyOn(window.localStorage.__proto__, "clear");
jest.spyOn(window.localStorage.__proto__, "getItem");
jest.spyOn(window.localStorage.__proto__, "setItem");
jest.spyOn(window.localStorage.__proto__, "removeItem");

(global as any).WebSocket = WebSocket;

(global as any).TextDecoder = TextDecoder;
(global as any).TextEncoder = TextEncoder;
(global as any).ResizeObserver = ResizeObserver;

// IntersectionObserver isn't available in test environment
const mockIntersectionObserver = jest.fn();
mockIntersectionObserver.mockReturnValue({
  observe: () => null,
  unobserve: () => null,
  disconnect: () => null,
});
window.IntersectionObserver = mockIntersectionObserver;

// Disable console.error/warn for some known test warnings

const originalError = console.error.bind(console.error);
const originalWarn = console.warn.bind(console.warn);

const ignoredMgs = [
  "Could not create web worker", // monaco uses web workers, but we don't need them for testing
  "MonacoEnvironment.getWorkerUrl or MonacoEnvironment.getWorker", // monaco uses web workers, but we don't need them for testing
  "react-tooltip.min.cjs", // react-tooltip complains about some tests without an "act()" wrapper
];

beforeAll(() => {
  console.error = msg => ignoredMgs.includes(msg.toString()) && originalError(msg);
  console.warn = msg => ignoredMgs.includes(msg.toString()) && originalWarn(msg);
});

afterAll(() => {
  console.error = originalError;
  console.warn = originalWarn;
});
