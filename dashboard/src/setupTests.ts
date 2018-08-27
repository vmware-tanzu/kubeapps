import "jest-enzyme";
import "raf/polyfill"; // polyfill for requestAnimationFrame

import { configure } from "enzyme";
import * as Adapter from "enzyme-adapter-react-16";
import { WebSocket } from "mock-socket";

configure({ adapter: new Adapter() });

// Mock browser specific APIs like localstorage or Websocket
const localStorageMock = {
  clear: jest.fn(),
  getItem: jest.fn(() => null),
  setItem: jest.fn(),
};

(global as any).localStorage = localStorageMock;
(global as any).WebSocket = WebSocket;
