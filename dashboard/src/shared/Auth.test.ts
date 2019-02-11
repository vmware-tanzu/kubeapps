import Axios from "axios";
import { Auth } from "./Auth";

describe("Auth", () => {
  beforeEach(() => {
    Axios.get = jest.fn();
  });
  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("fetchToken", () => {
    it("should fetch a token", async () => {
      Axios.get = jest.fn(() => {
        return { headers: { authorization: "foo" } };
      });
      const token = await Auth.fetchToken();
      expect(token).toEqual("foo");
    });
  });
  it("should catch errors in the request", async () => {
    Axios.get = jest.fn(() => {
      throw new Error();
    });
    const token = await Auth.fetchToken();
    expect(token).toEqual(null);
  });
});
