import { LOCATION_CHANGE, RouterActionType } from "connected-react-router";
import context from "jest-plugin-context";
import namespaceReducer from "./namespace";

describe("namespaceReducer", () => {
  const location = {
    hash: "",
    search: "",
    state: "",
  };
  const initialState = namespaceReducer(undefined, {} as any);

  context("when LOCATION CHANGE", () => {
    it("changes the current stored namespace if it is in the URL", () => {
      const testCases = [
        { path: "/ns/cyberdyne/apps", current: "cyberdyne" },
        { path: "/cyberdyne/apps", current: "default" },
        { path: "/ns/T-600/charts", current: "T-600" },
      ];
      testCases.forEach(tc => {
        expect(
          namespaceReducer(undefined, {
            type: LOCATION_CHANGE,
            payload: {
              location: { ...location, pathname: tc.path },
              action: "PUSH" as RouterActionType,
            },
          }),
        ).toEqual({ ...initialState, current: tc.current });
      });
    });
  });
});
