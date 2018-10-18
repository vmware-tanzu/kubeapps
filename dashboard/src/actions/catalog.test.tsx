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
const clusterClass = { metadata: { name: "cluster-class" } } as any;

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
let boomFn: any;
const errorPayload = (op: string) => ({ err: new Error("Boom!"), op });

beforeEach(() => {
  store = mockStore();

  ServiceInstance.create = jest.fn().mockImplementationOnce(() => {
    return { metadata: { name: testArgs.instanceName } };
  });
  ServiceInstance.list = jest.fn().mockImplementationOnce(() => {
    return [serviceInstance];
  });
  ServiceBinding.create = jest.fn().mockImplementationOnce(() => {
    return { metadata: { name: testArgs.bindingName } };
  });
  ServiceBinding.delete = jest.fn();
  ServiceBinding.list = jest.fn().mockImplementationOnce(() => {
    return [bindingWithSecret];
  });
  ServiceCatalog.getServiceBrokers = jest.fn().mockImplementationOnce(() => {
    return [broker];
  });
  ServiceCatalog.getServiceClasses = jest.fn().mockImplementationOnce(() => {
    return [clusterClass];
  });
  ServiceCatalog.getServicePlans = jest.fn().mockImplementationOnce(() => {
    return [servicePlan];
  });
  ServiceCatalog.deprovisionInstance = jest.fn();
  ServiceCatalog.syncBroker = jest.fn();
  boomFn = jest.fn().mockImplementationOnce(() => {
    throw new Error("Boom!");
  });
  ServiceCatalog.isCatalogInstalled = jest.fn().mockImplementationOnce(() => true);
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
    ServiceInstance.create = boomFn;

    const expectedActions = [
      {
        type: getType(catalogActions.errorCatalog),
        payload: errorPayload("create"),
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("filters the submitted parameters if they are empty", async () => {
    const params = {
      a: 1,
      f: {
        g: [],
        h: {
          i: "",
        },
      },
      j: {
        k: {
          l: "m",
        },
      },
    };
    // It omits "f" because it's empty but not "a" or "j"
    const expectedParams = { a: 1, j: { k: { l: "m" } } };
    const cmd = catalogActions.provision(
      testArgs.releaseName,
      testArgs.namespace,
      testArgs.className,
      testArgs.planName,
      params,
    );
    await store.dispatch(cmd);
    expect(ServiceInstance.create).toHaveBeenCalledWith(
      testArgs.releaseName,
      testArgs.namespace,
      testArgs.className,
      testArgs.planName,
      expectedParams,
    );
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
    ServiceCatalog.deprovisionInstance = boomFn;

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
    ServiceBinding.create = boomFn;

    const expectedActions = [
      {
        type: getType(catalogActions.errorCatalog),
        payload: errorPayload("create"),
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("filters the submitted parameters if they are empty", async () => {
    const params = {
      a: 1,
      b: {
        c: [],
      },
      f: {
        g: [],
        h: {
          i: "",
        },
      },
      j: {
        k: {
          l: "m",
        },
      },
    };
    // It omits "f" because it's empty but not "a" or "j"
    const expectedParams = { a: 1, j: { k: { l: "m" } } };
    const cmd = catalogActions.addBinding(
      testArgs.bindingName,
      testArgs.instanceName,
      testArgs.namespace,
      params,
    );
    await store.dispatch(cmd);
    expect(ServiceBinding.create).toHaveBeenCalledWith(
      testArgs.bindingName,
      testArgs.instanceName,
      testArgs.namespace,
      expectedParams,
    );
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
    ServiceBinding.delete = boomFn;

    const expectedActions = [
      {
        type: getType(catalogActions.errorCatalog),
        payload: errorPayload("delete"),
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("sync", () => {
  const provisionCMD = catalogActions.sync(broker);

  it("calls ServiceCatalog.syncBroker if no error", async () => {
    await store.dispatch(provisionCMD);
    expect(store.getActions().length).toBe(0);
    expect(ServiceCatalog.syncBroker).toHaveBeenCalledWith(broker);
  });

  it("dispatches errorCatalog if error", async () => {
    ServiceCatalog.syncBroker = boomFn;

    const expectedActions = [
      {
        type: getType(catalogActions.errorCatalog),
        payload: { err: new Error("Boom!"), op: "update" },
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("getBindings", () => {
  const provisionCMD = catalogActions.getBindings(testArgs.namespace);

  it("calls ServiceBinding.list and dispatches binding actions if no error", async () => {
    const expectedActions = [
      {
        type: getType(catalogActions.requestBindingsWithSecrets),
      },
      {
        type: getType(catalogActions.receiveBindingsWithSecrets),
        payload: [bindingWithSecret],
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
    expect(ServiceBinding.list).toHaveBeenCalledWith(testArgs.namespace);
  });

  it("dispatches requestBindingsWithSecrets and errorCatalog if error", async () => {
    ServiceBinding.list = boomFn;

    const expectedActions = [
      {
        type: getType(catalogActions.requestBindingsWithSecrets),
      },
      {
        type: getType(catalogActions.errorCatalog),
        payload: errorPayload("fetch"),
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("getBrokers", () => {
  const provisionCMD = catalogActions.getBrokers();

  it("calls ServiceCatalog.getServiceBrokers and dispatches requestBrokers and receiveBroker if no error", async () => {
    const expectedActions = [
      {
        type: getType(catalogActions.requestBrokers),
      },
      {
        type: getType(catalogActions.receiveBrokers),
        payload: [broker],
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
    expect(ServiceCatalog.getServiceBrokers).toHaveBeenCalled();
  });

  it("dispatches requestBrokers and errorCatalog if error", async () => {
    ServiceCatalog.getServiceBrokers = boomFn;

    const expectedActions = [
      {
        type: getType(catalogActions.requestBrokers),
      },
      {
        type: getType(catalogActions.errorCatalog),
        payload: errorPayload("fetch"),
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("getClasses", () => {
  const provisionCMD = catalogActions.getClasses();

  it("calls ServiceCatalog.getServiceClasses and dispatches requestClasses and receiveClasses if no error", async () => {
    const expectedActions = [
      {
        type: getType(catalogActions.requestClasses),
      },
      {
        type: getType(catalogActions.receiveClasses),
        payload: [clusterClass],
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
    expect(ServiceCatalog.getServiceClasses).toHaveBeenCalled();
  });

  it("dispatches requestClasses and errorCatalog if error", async () => {
    ServiceCatalog.getServiceClasses = boomFn;

    const expectedActions = [
      {
        type: getType(catalogActions.requestClasses),
      },
      {
        type: getType(catalogActions.errorCatalog),
        payload: errorPayload("fetch"),
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("getInstances", () => {
  const provisionCMD = catalogActions.getInstances(testArgs.namespace);

  it("calls ServiceInstance.list and dispatches requestInstances and receiveInstances if no error", async () => {
    const expectedActions = [
      {
        type: getType(catalogActions.requestInstances),
      },
      {
        type: getType(catalogActions.receiveInstances),
        payload: [serviceInstance],
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
    expect(ServiceInstance.list).toHaveBeenCalledWith(testArgs.namespace);
  });

  it("dispatches requestInstances and errorCatalog if error", async () => {
    ServiceInstance.list = boomFn;

    const expectedActions = [
      {
        type: getType(catalogActions.requestInstances),
      },
      {
        type: getType(catalogActions.errorCatalog),
        payload: errorPayload("fetch"),
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("getPlans", () => {
  const provisionCMD = catalogActions.getPlans();

  it("calls ServiceCatalog.getServicePlans and dispatches requestPlans and receivePlans if no error", async () => {
    const expectedActions = [
      {
        type: getType(catalogActions.requestPlans),
      },
      {
        type: getType(catalogActions.receivePlans),
        payload: [servicePlan],
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
    expect(ServiceCatalog.getServicePlans).toHaveBeenCalled();
  });

  it("dispatches requestPlans and errorCatalog if error", async () => {
    ServiceCatalog.getServicePlans = boomFn;

    const expectedActions = [
      {
        type: getType(catalogActions.requestPlans),
      },
      {
        type: getType(catalogActions.errorCatalog),
        payload: errorPayload("fetch"),
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("checkCatalogInstalled", () => {
  const provisionCMD = catalogActions.checkCatalogInstalled();

  it("dispatches installed = true if installed", async () => {
    const expectedActions = [
      {
        type: getType(catalogActions.installed),
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
    expect(ServiceCatalog.isCatalogInstalled).toHaveBeenCalled();
  });

  it("dispatches installed = false otherwise", async () => {
    ServiceCatalog.isCatalogInstalled = jest.fn().mockImplementationOnce(() => false);

    const expectedActions = [
      {
        type: getType(catalogActions.notInstalled),
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
    expect(ServiceCatalog.isCatalogInstalled).toHaveBeenCalled();
  });
});
