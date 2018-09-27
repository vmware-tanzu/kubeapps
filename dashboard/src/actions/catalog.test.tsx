import context from "jest-plugin-context";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";
import actions from ".";
import { ServiceBinding } from "../shared/ServiceBinding";
import { IServiceBroker, IServicePlan, ServiceCatalog } from "../shared/ServiceCatalog";
import { IServiceInstance, ServiceInstance } from "../shared/ServiceInstance";

const { catalog: catalogActions } = actions;

const mockStore = configureMockStore([thunk]);
const broker = { metadata: { name: "wall-street" } } as IServiceBroker;
const servicePlan = { metadata: { name: "bubble-it-up" } } as IServicePlan;
const serviceInstance = { metadata: { name: "25-years-morgage" } } as IServiceInstance;
const bindingWithSecret = { binding: "binding", secret: "secret" } as any;

let store: any;
const testArgs = {
  releaseName: "my-release",
  namespace: "my-namespace",
  className: "my-class",
  planName: "myPlan",
  bindingName: "my-binding",
  params: {},
  instanceName: "my-instance",
};

beforeEach(() => {
  store = mockStore();

  ServiceInstance.create = jest.fn().mockImplementationOnce(() => {
    return { metadata: { name: testArgs.instanceName } };
  });
  ServiceBinding.create = jest.fn().mockImplementationOnce(() => {
    return { metadata: { name: testArgs.bindingName } };
  });
  ServiceBinding.delete = jest.fn();
  ServiceCatalog.deprovisionInstance = jest.fn();
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

// Async action creators
describe("provision", () => {
  const provisionCMD = catalogActions.provision(
    testArgs.releaseName,
    testArgs.namespace,
    testArgs.className,
    testArgs.planName,
    testArgs.params,
  );

  it("calls ServiceInstance.create and returns true if no error", async () => {
    const res = await store.dispatch(provisionCMD);
    expect(res).toBe(true);

    expect(store.getActions().length).toBe(0);
    expect(ServiceInstance.create).toHaveBeenCalledWith(
      testArgs.releaseName,
      testArgs.namespace,
      testArgs.className,
      testArgs.planName,
      {},
    );
  });

  it("dispatches errorCatalog if error creating the instance", async () => {
    ServiceInstance.create = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      {
        type: getType(catalogActions.errorCatalog),
        payload: { err: new Error("Boom!"), op: "create" },
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("deprovision", () => {
  const provisionCMD = catalogActions.deprovision(serviceInstance);

  it("calls ServiceInstance.deprovisionInstance and returns true if no error", async () => {
    const res = await store.dispatch(provisionCMD);
    expect(res).toBe(true);

    expect(store.getActions().length).toBe(0);
    expect(ServiceCatalog.deprovisionInstance).toHaveBeenCalledWith(serviceInstance);
  });

  it("dispatches errorCatalog if error", async () => {
    ServiceCatalog.deprovisionInstance = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      {
        type: getType(catalogActions.errorCatalog),
        payload: { err: new Error("Boom!"), op: "deprovision" },
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("addBinding", () => {
  const provisionCMD = catalogActions.addBinding(
    testArgs.bindingName,
    testArgs.instanceName,
    testArgs.namespace,
    testArgs.params,
  );

  it("calls ServiceBinding.create and returns true if no error", async () => {
    const res = await store.dispatch(provisionCMD);
    expect(res).toBe(true);

    expect(store.getActions().length).toBe(0);
    expect(ServiceBinding.create).toHaveBeenCalledWith(
      testArgs.bindingName,
      testArgs.instanceName,
      testArgs.namespace,
      testArgs.params,
    );
  });

  it("dispatches errorCatalog if error", async () => {
    ServiceBinding.create = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      {
        type: getType(catalogActions.errorCatalog),
        payload: { err: new Error("Boom!"), op: "create" },
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("removeBinding", () => {
  const provisionCMD = catalogActions.removeBinding(testArgs.bindingName, testArgs.namespace);

  it("calls ServiceBinding.delete and returns true if no error", async () => {
    const res = await store.dispatch(provisionCMD);
    expect(res).toBe(true);

    expect(store.getActions().length).toBe(0);
    expect(ServiceBinding.delete).toHaveBeenCalledWith(testArgs.bindingName, testArgs.namespace);
  });

  it("dispatches errorCatalog if error", async () => {
    ServiceBinding.delete = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      {
        type: getType(catalogActions.errorCatalog),
        payload: { err: new Error("Boom!"), op: "delete" },
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });
});
