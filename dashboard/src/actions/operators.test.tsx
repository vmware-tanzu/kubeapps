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

describe("getOperators", () => {
  it("returns an ordered list of operators based on the name", async () => {
    Operators.getOperators = jest.fn(() => [
      { metadata: { name: "foo" } },
      { metadata: { name: "bar" } },
    ]);
    const sortedOperators = [{ metadata: { name: "bar" } }, { metadata: { name: "foo" } }];
    const expectedActions = [
      {
        type: getType(operatorActions.requestOperators),
      },
      {
        type: getType(operatorActions.receiveOperators),
        payload: sortedOperators,
      },
    ];
    await store.dispatch(operatorActions.getOperators("default"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error", async () => {
    Operators.getOperators = jest.fn(() => {
      throw new Error("Boom!");
    });
    const expectedActions = [
      {
        type: getType(operatorActions.requestOperators),
      },
      {
        type: getType(operatorActions.errorOperators),
        payload: new Error("Boom!"),
      },
    ];
    await store.dispatch(operatorActions.getOperators("default"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("getOperator", () => {
  it("returns an an operator", async () => {
    const op = { metadata: { name: "foo" } };
    Operators.getOperator = jest.fn(() => op);
    const expectedActions = [
      {
        type: getType(operatorActions.requestOperator),
      },
      {
        type: getType(operatorActions.receiveOperator),
        payload: op,
      },
    ];
    await store.dispatch(operatorActions.getOperator("default", "foo"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error", async () => {
    Operators.getOperator = jest.fn(() => {
      throw new Error("Boom!");
    });
    const expectedActions = [
      {
        type: getType(operatorActions.requestOperator),
      },
      {
        type: getType(operatorActions.errorOperators),
        payload: new Error("Boom!"),
      },
    ];
    await store.dispatch(operatorActions.getOperator("default", "foo"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("getCSVs", () => {
  it("returns an ordered list of csvs based on the name", async () => {
    Operators.getCSVs = jest.fn(() => [
      { metadata: { name: "foo" } },
      { metadata: { name: "bar" } },
    ]);
    const sortedCSVs = [{ metadata: { name: "bar" } }, { metadata: { name: "foo" } }];
    const expectedActions = [
      {
        type: getType(operatorActions.requestCSVs),
      },
      {
        type: getType(operatorActions.receiveCSVs),
        payload: sortedCSVs,
      },
    ];
    await store.dispatch(operatorActions.getCSVs("default"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error", async () => {
    Operators.getCSVs = jest.fn(() => {
      throw new Error("Boom!");
    });
    const expectedActions = [
      {
        type: getType(operatorActions.requestCSVs),
      },
      {
        type: getType(operatorActions.errorCSVs),
        payload: new Error("Boom!"),
      },
    ];
    await store.dispatch(operatorActions.getCSVs("default"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});
