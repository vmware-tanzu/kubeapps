import Axios from "axios";
import { Auth } from "./Auth";

describe("Auth", () => {
  beforeEach(() => {
    Axios.get = jest.fn();
  });
  afterEach(() => {
    jest.resetAllMocks();
  });
  it("should get an URL with the given token", async () => {
    const mock = jest.fn();
    Axios.get = mock;
    await Auth.validateToken("foo");
    expect(mock.mock.calls[0]).toEqual(["api/kube/", { headers: { Authorization: "Bearer foo" } }]);
  });

  describe("when there is an error", () => {
    it("should throw error for 401 error codes", async () => {
      const mock = jest.fn().mockRejectedValue({ response: { status: 401 } });
      Axios.get = mock;
      try {
        await Auth.validateToken("foo");
      } catch (e) {
        expect(e.message).toEqual("invalid token");
      }
    });
  });

  describe("fetchOIDCToken", () => {
    it("should fetch a token", async () => {
      Axios.head = jest.fn(path => {
        expect(path).toEqual("");
        return { headers: { authorization: "Bearer foo" } };
      });
      const token = await Auth.fetchOIDCToken();
      expect(token).toEqual("foo");
    });
  });
  it("should not return a token if the info is not present", async () => {
    Axios.head = jest.fn(() => {
      return {};
    });
    const token = await Auth.fetchOIDCToken();
    expect(token).toEqual(null);
  });
  it("should not return a token if the call fails", async () => {
    Axios.head = jest.fn(() => {
      throw new Error();
    });
    const token = await Auth.fetchOIDCToken();
    expect(token).toEqual(null);
  });
});
