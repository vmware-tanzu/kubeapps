import context from "jest-plugin-context";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";
import actions from ".";
import { IServiceBindingWithSecret } from "../shared/ServiceBinding";
import { IServiceBroker, IServicePlan } from "../shared/ServiceCatalog";
import { IServiceInstance } from "../shared/ServiceInstance";

const { catalog: catalogActions } = actions;

const mockStore = configureMockStore([thunk]);
const broker = { metadata: { name: "wall-street" } } as IServiceBroker;
const servicePlan = { metadata: { name: "bubble-it-up" } } as IServicePlan;
const serviceInstance = { metadata: { name: "25-years-morgage" } } as IServiceInstance;
const bindingWithSecret = { binding: "binding", secret: "secret" } as any;

let store: any;

beforeEach(() => {
  store = mockStore();
});

// Regular action creators
interface ITestCase {
  name: string;
  action: (...args: any[]) => any;
  args?: any;
  payload?: any;
}

const actionTestCases: ITestCase[] = [
  { name: "checkCatalogInstall", action: catalogActions.checkCatalogInstall },
  { name: "installed", action: catalogActions.installed },
  { name: "notInstalled", action: catalogActions.notInstalled },
  { name: "requestBrokers", action: catalogActions.requestBrokers },
  {
    name: "receiveBrokers",
    action: catalogActions.receiveBrokers,
    args: [[broker]],
    payload: [broker],
  },
  { name: "requestPlans", action: catalogActions.requestPlans },
  {
    name: "receivePlans",
    action: catalogActions.receivePlans,
    args: [[servicePlan]],
    payload: [servicePlan],
  },
  { name: "requestInstances", action: catalogActions.requestInstances },
  {
    name: "receiveInstances",
    action: catalogActions.receiveInstances,
    args: [[serviceInstance]],
    payload: [serviceInstance],
  },
  { name: "requestBindingsWithSecrets", action: catalogActions.requestBindingsWithSecrets },
  {
    name: "receiveBindingsWithSecrets",
    action: catalogActions.receiveBindingsWithSecrets,
    args: [[bindingWithSecret]],
    payload: [bindingWithSecret],
  },
  { name: "requestClasses", action: catalogActions.requestClasses },
  {
    name: "receiveClasses",
    action: catalogActions.receiveClasses,
    args: [[bindingWithSecret]],
    payload: [bindingWithSecret],
  },
  {
    name: "errorCatalog",
    action: catalogActions.errorCatalog,
    args: [new Error("foo"), "create"],
    payload: { err: new Error("foo"), op: "create" },
  },
];

actionTestCases.forEach(tc => {
  describe(tc.name, () => {
    it("has expected structure", () => {
      const actionResult = tc.action.call(null, ...tc.args);
      expect(actionResult).toEqual({
        type: getType(tc.action),
        payload: tc.payload,
      });
    });
  });
});
