import { getType } from "typesafe-actions";
import actions from "../actions";
import { IKubeState } from "../shared/types";
import { IKubeItem } from "./../shared/types";
import kubeReducer from "./kube";

describe("authReducer", () => {
  let initialState: IKubeState;

  const actionTypes = {
    requestResource: getType(actions.kube.requestResource),
    receiveResource: getType(actions.kube.receiveResource),
    errorKube: getType(actions.kube.errorKube),
  };

  beforeEach(() => {
    initialState = {
      items: {},
    };
  });

  describe("reducer actions", () => {
    it("stores item", () => {
      const payload = { foo: {} as IKubeItem };
      [
        actionTypes.requestResource as any,
        actionTypes.receiveResource as any,
        actionTypes.errorKube as any,
      ].forEach(type => {
        expect(
          kubeReducer(undefined, {
            type,
            payload,
          }),
        ).toEqual({ ...initialState, items: { foo: {} as IKubeItem } });
      });
    });
  });
});
