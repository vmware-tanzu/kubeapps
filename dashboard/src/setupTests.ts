import "jest-enzyme";
import "raf/polyfill"; // polyfill for requestAnimationFrame

import { configure } from "enzyme";
import Adapter from "enzyme-adapter-react-16";
import { WebSocket } from "mock-socket";

configure({ adapter: new Adapter() });

// Mock browser specific APIs like localstorage or Websocket
jest.spyOn(window.localStorage.__proto__, "clear");
jest.spyOn(window.localStorage.__proto__, "getItem");
jest.spyOn(window.localStorage.__proto__, "setItem");
jest.spyOn(window.localStorage.__proto__, "removeItem");

(global as any).WebSocket = WebSocket;
