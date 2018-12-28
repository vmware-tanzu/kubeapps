import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";

import actions from ".";
import { Kube } from "../shared/Kube";
import { IKubeState } from "../shared/types";

const mockStore = configureMockStore([thunk]);

let store: any;

beforeEach(() => {
  const state: IKubeState = {
    items: {},
  };
  store = mockStore({
    kube: {
      state,
    },
  });
});

describe("get resources", () => {
  const getResourceOrig = Kube.getResource;
  let getResourceMock: jest.Mock;
  beforeEach(() => {
    getResourceMock = jest.fn(() => []);
    Kube.getResource = getResourceMock;
  });
  afterEach(() => {
    Kube.getResource = getResourceOrig;
  });
  it("fetches a resource", async () => {
    const expectedActions = [
      {
        type: getType(actions.kube.requestResource),
        payload: "api/kube/api/v1/namespaces/default/pods/foo",
      },
      {
        type: getType(actions.kube.receiveResource),
        payload: {
          key: "api/kube/api/v1/namespaces/default/pods/foo",
          resource: [],
        },
      },
    ];
    await store.dispatch(actions.kube.getResource("v1", "pods", "default", "foo"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(getResourceMock.mock.calls[0]).toEqual(["v1", "pods", "default", "foo", undefined]);
  });
});
