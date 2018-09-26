import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";
import actions from ".";
import Config from "../shared/Config";

const mockStore = configureMockStore([thunk]);

let store: any;
const CONFIG = "have you tried to turn it off and on again";

beforeEach(() => {
  Config.getConfig = jest.fn(() => CONFIG);

  store = mockStore();
});

describe("getConfig", () => {
  it("dispatches request config and its returned value", async () => {
    const expectedActions = [
      {
        type: getType(actions.config.requestConfig),
      },
      {
        payload: CONFIG,
        type: getType(actions.config.receiveConfig),
      },
    ];

    await store.dispatch(actions.config.getConfig());
    expect(store.getActions()).toEqual(expectedActions);
  });
});
