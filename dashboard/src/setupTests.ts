// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Adapter from "@wojtekmaj/enzyme-adapter-react-17";
import Enzyme from "enzyme";
import "jest-enzyme";
import { WebSocket } from "mock-socket";
import { TextDecoder, TextEncoder } from "util";
import ResizeObserver from "resize-observer-polyfill";

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
