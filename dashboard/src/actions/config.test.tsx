import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import Config, { SupportedThemes } from "shared/Config";
import { getType } from "typesafe-actions";
import actions from ".";

const mockStore = configureMockStore([thunk]);

let store: any;
const testConfig = "have you tried to turn it off and on again";

beforeEach(() => {
  Config.getConfig = jest.fn().mockReturnValue(testConfig);

  store = mockStore();
});

afterEach(() => {
  jest.restoreAllMocks();
});

describe("getConfig", () => {
  it("dispatches request config and its returned value", async () => {
    const expectedActions = [
      {
        type: getType(actions.config.requestConfig),
      },
      {
        payload: testConfig,
        type: getType(actions.config.receiveConfig),
      },
    ];

    await store.dispatch(actions.config.getConfig());
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("getTheme", () => {
  it("dispatches request config and its returned value", async () => {
    Config.getTheme = jest.fn().mockReturnValue(SupportedThemes.dark);
    const expectedActions = [
      {
        payload: SupportedThemes.dark,
        type: getType(actions.config.receiveTheme),
      },
    ];

    await store.dispatch(actions.config.getTheme());
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("setUserTheme", () => {
  it("dispatches request config and its returned value", async () => {
    Config.setUserTheme = jest.fn();
    const expectedActions = [
      {
        payload: SupportedThemes.dark,
        type: getType(actions.config.receiveTheme),
      },
    ];

    await store.dispatch(actions.config.setUserTheme(SupportedThemes.dark));
    expect(store.getActions()).toEqual(expectedActions);
  });
});
