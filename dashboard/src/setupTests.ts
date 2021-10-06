import Adapter from "@wojtekmaj/enzyme-adapter-react-17";
import Enzyme from "enzyme";
import "jest-enzyme";
import { WebSocket } from "mock-socket";

Enzyme.configure({ adapter: new Adapter() });

// Mock browser specific APIs like localstorage or Websocket
jest.spyOn(window.localStorage.__proto__, "clear");
jest.spyOn(window.localStorage.__proto__, "getItem");
jest.spyOn(window.localStorage.__proto__, "setItem");
jest.spyOn(window.localStorage.__proto__, "removeItem");

(global as any).WebSocket = WebSocket;
