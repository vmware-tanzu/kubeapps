import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";
import actions from ".";
import Config, { SupportedThemes } from "../shared/Config";

const mockStore = configureMockStore([thunk]);

let store: any;
const testConfig = "have you tried to turn it off and on again";

beforeEach(() => {
  Config.getConfig = jest.fn().mockReturnValue(testConfig);

  store = mockStore();
});

afterEach(() => {
  jest.resetAllMocks();
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

describe("setTheme", () => {
  it("dispatches request config and its returned value", async () => {
    Config.setTheme = jest.fn();
    const expectedActions = [
      {
        payload: SupportedThemes.dark,
        type: getType(actions.config.setThemeState),
      },
    ];

    await store.dispatch(actions.config.setTheme(SupportedThemes.dark));
    expect(store.getActions()).toEqual(expectedActions);
  });
});
