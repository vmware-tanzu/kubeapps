import { getType } from "typesafe-actions";
import actions from "../actions";
import { IKubeState, IResource } from "../shared/types";
import kubeReducer from "./kube";

describe("authReducer", () => {
  let initialState: IKubeState;

  const actionTypes = {
    requestResource: getType(actions.kube.requestResource),
    receiveResource: getType(actions.kube.receiveResource),
    errorKube: getType(actions.kube.receiveResourceError),
  };

  beforeEach(() => {
    initialState = {
      items: {},
    };
  });

  describe("reducer actions", () => {
    it("request an item", () => {
      const payload = "foo";
      const type = actionTypes.requestResource as any;
      expect(kubeReducer(undefined, { type, payload })).toEqual({
        ...initialState,
        items: { foo: { isFetching: true } },
      });
    });
    it("receives an item", () => {
      const payload = { key: "foo", resource: { metadata: { name: "foo" } } as IResource };
      const type = actionTypes.receiveResource as any;
      expect(kubeReducer(undefined, { type, payload })).toEqual({
        ...initialState,
        items: { foo: { isFetching: false, item: { metadata: { name: "foo" } } } },
      });
    });
    it("receives an error", () => {
      const error = new Error("bar");
      const payload = { key: "foo", error };
      const type = actionTypes.errorKube as any;
      expect(kubeReducer(undefined, { type, payload })).toEqual({
        ...initialState,
        items: { foo: { isFetching: false, error } },
      });
    });
  });
});
