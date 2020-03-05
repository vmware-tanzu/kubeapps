import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";
import actions from ".";
import { Operators } from "../shared/Operators";

const { operators: operatorActions } = actions;
const mockStore = configureMockStore([thunk]);

let store: any;

beforeEach(() => {
  store = mockStore({});
  Operators.isOLMInstalled = jest.fn();
});

afterEach(jest.resetAllMocks);

describe("checkOLMInstalled", () => {
  it("dispatches OLM_INSTALLED when succeded", async () => {
    Operators.isOLMInstalled = jest.fn(() => true);
    const expectedActions = [
      {
        type: getType(operatorActions.checkingOLM),
      },
      {
        type: getType(operatorActions.OLMInstalled),
      },
    ];
    await store.dispatch(operatorActions.checkOLMInstalled());
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches OLM_NOT_INSTALLED when failed", async () => {
    Operators.isOLMInstalled = jest.fn(() => false);
    const expectedActions = [
      {
        type: getType(operatorActions.checkingOLM),
      },
      {
        type: getType(operatorActions.OLMNotInstalled),
      },
    ];
    await store.dispatch(operatorActions.checkOLMInstalled());
    expect(store.getActions()).toEqual(expectedActions);
  });
});
