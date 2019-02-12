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
    await Auth.validate("foo");
    expect(mock.mock.calls[0]).toEqual(["api/kube/", { headers: { Authorization: "Bearer foo" } }]);
  });
  it("should skip the token if not given", async () => {
    const mock = jest.fn();
    Axios.get = mock;
    await Auth.validate();
    expect(mock.mock.calls[0]).toEqual(["api/kube/", {}]);
  });

  describe("when there is an error", () => {
    it("should throw errors for 401 and 403 error codes", () => {
      [401, 402].forEach(async code => {
        const mock = jest.fn().mockRejectedValue({ response: { status: code } });
        Axios.get = mock;
        try {
          await Auth.validate("foo");
        } catch (e) {
          expect(e.message).toEqual("invalid token");
        }
      });
    });
  });
});
