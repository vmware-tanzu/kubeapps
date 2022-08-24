// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { Operators } from "shared/Operators";
import { IResource, IStoreState } from "shared/types";
import { getType } from "typesafe-actions";
import actions from ".";

const { operators: operatorActions } = actions;
const mockStore = configureMockStore([thunk]);

let store: any;

beforeEach(() => {
  store = mockStore({} as Partial<IStoreState>);
});

afterEach(jest.restoreAllMocks);

describe("checkOLMInstalled", () => {
  it("dispatches OLM_INSTALLED when succeeded", async () => {
    Operators.isOLMInstalled = jest.fn().mockReturnValue(true);
    const expectedActions = [
      {
        type: getType(operatorActions.checkingOLM),
      },
      {
        type: getType(operatorActions.OLMInstalled),
      },
    ];
    await store.dispatch(operatorActions.checkOLMInstalled("default", "ns"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches error", async () => {
    Operators.isOLMInstalled = jest.fn(() => {
      throw new Error("nope");
    });
    const expectedActions = [
      {
        type: getType(operatorActions.checkingOLM),
      },
      {
        type: getType(operatorActions.errorOLMCheck),
        payload: new Error("nope"),
      },
    ];
    await store.dispatch(operatorActions.checkOLMInstalled("default", "ns"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("getOperators", () => {
  it("returns an ordered list of operators based on the name", async () => {
    Operators.getOperators = jest
      .fn()
      .mockResolvedValue([{ metadata: { name: "foo" } }, { metadata: { name: "bar" } }]);
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
    await store.dispatch(operatorActions.getOperators("default", "default"));
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
    await store.dispatch(operatorActions.getOperators("default", "default"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("getOperator", () => {
  it("returns an operator", async () => {
    const op = { metadata: { name: "foo" } };
    Operators.getOperator = jest.fn().mockReturnValue(op);
    const expectedActions = [
      {
        type: getType(operatorActions.requestOperator),
      },
      {
        type: getType(operatorActions.receiveOperator),
        payload: op,
      },
    ];
    await store.dispatch(operatorActions.getOperator("default", "default", "foo"));
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
    await store.dispatch(operatorActions.getOperator("default", "default", "foo"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("getCSVs", () => {
  it("returns an ordered list of csvs based on the name", async () => {
    Operators.getCSVs = jest
      .fn()
      .mockResolvedValue([{ metadata: { name: "foo" } }, { metadata: { name: "bar" } }]);
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
    await store.dispatch(operatorActions.getCSVs("default", "default"));
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
    await store.dispatch(operatorActions.getCSVs("default", "default"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("getCSV", () => {
  it("returns a ClusterServiceVersion", async () => {
    const csv = { metadata: { name: "foo" } };
    Operators.getCSV = jest.fn().mockReturnValue(csv);
    const expectedActions = [
      {
        type: getType(operatorActions.requestCSV),
      },
      {
        type: getType(operatorActions.receiveCSV),
        payload: csv,
      },
    ];
    await store.dispatch(operatorActions.getCSV("default", "default", "foo"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error", async () => {
    Operators.getCSV = jest.fn(() => {
      throw new Error("Boom!");
    });
    const expectedActions = [
      {
        type: getType(operatorActions.requestCSV),
      },
      {
        type: getType(operatorActions.errorCSVs),
        payload: new Error("Boom!"),
      },
    ];
    await store.dispatch(operatorActions.getCSV("default", "default", "foo"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("createResource", () => {
  it("creates a resource", async () => {
    const resource = {} as IResource;
    Operators.createResource = jest.fn().mockReturnValue(resource);
    const expectedActions = [
      {
        type: getType(operatorActions.creatingResource),
      },
      {
        type: getType(operatorActions.resourceCreated),
        payload: resource,
      },
    ];
    await store.dispatch(operatorActions.createResource("default", "default", "v1", "pods", {}));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error", async () => {
    Operators.createResource = jest.fn(() => {
      throw new Error("Boom!");
    });
    const expectedActions = [
      {
        type: getType(operatorActions.creatingResource),
      },
      {
        type: getType(operatorActions.errorResourceCreate),
        payload: new Error("Boom!"),
      },
    ];
    await store.dispatch(operatorActions.createResource("default", "default", "v1", "pods", {}));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("updateResource", () => {
  it("updates a resource", async () => {
    const resource = {} as IResource;
    Operators.updateResource = jest.fn().mockReturnValue(resource);
    const expectedActions = [
      {
        type: getType(operatorActions.updatingResource),
      },
      {
        type: getType(operatorActions.resourceUpdated),
        payload: resource,
      },
    ];
    await store.dispatch(
      operatorActions.updateResource("default", "default", "v1", "pods", "foo", {}),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error", async () => {
    Operators.updateResource = jest.fn(() => {
      throw new Error("Boom!");
    });
    const expectedActions = [
      {
        type: getType(operatorActions.updatingResource),
      },
      {
        type: getType(operatorActions.errorResourceUpdate),
        payload: new Error("Boom!"),
      },
    ];
    await store.dispatch(
      operatorActions.updateResource("default", "default", "v1", "pods", "foo", {}),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("listResources", () => {
  it("list resources in a namespace", async () => {
    const csv = {
      metadata: { name: "foo", namespace: "default" },
      spec: {
        customresourcedefinitions: { owned: [{ name: "foo.kubeapps.com", version: "v1alpha1" }] },
      },
    };
    const resource = { metadata: { name: "resource" } };
    Operators.getCSVs = jest.fn().mockReturnValue([csv]);
    Operators.listResources = jest.fn().mockReturnValue({
      items: [resource],
    });
    const expectedActions = [
      {
        type: getType(operatorActions.requestCustomResources),
      },
      {
        type: getType(operatorActions.requestCSVs),
      },
      {
        type: getType(operatorActions.receiveCSVs),
        payload: [csv],
      },
      {
        type: getType(operatorActions.receiveCustomResources),
        payload: [resource],
      },
    ];
    await store.dispatch(operatorActions.getResources("default", "default"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(Operators.listResources).toHaveBeenCalledWith(
      "default",
      "default",
      "kubeapps.com/v1alpha1",
      "foo",
    );
  });

  it("dispatches an error if listing resources fail", async () => {
    const csv = {
      metadata: { name: "foo" },
      spec: {
        customresourcedefinitions: { owned: [{ name: "foo.kubeapps.com", version: "v1alpha1" }] },
      },
    };
    Operators.getCSVs = jest.fn().mockReturnValue([csv]);
    Operators.listResources = jest.fn(() => {
      throw new Error("Boom!");
    });
    const expectedActions = [
      {
        type: getType(operatorActions.requestCustomResources),
      },
      {
        type: getType(operatorActions.requestCSVs),
      },
      {
        type: getType(operatorActions.receiveCSVs),
        payload: [csv],
      },
      {
        type: getType(operatorActions.errorCustomResource),
        payload: new Error("Boom!"),
      },
      {
        type: getType(operatorActions.receiveCustomResources),
        payload: [],
      },
    ];
    await store.dispatch(operatorActions.getResources("default", "default"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("lists resources with their corresponding namespace when empty namespace is requested", async () => {
    const csvs = [
      {
        metadata: { name: "foo", namespace: "foo1-ns" },
        spec: {
          customresourcedefinitions: { owned: [{ name: "foo.kubeapps.com", version: "v1alpha1" }] },
        },
      },
      {
        metadata: { name: "foo", namespace: "foo2-ns" },
        spec: {
          customresourcedefinitions: { owned: [{ name: "foo.kubeapps.com", version: "v1alpha1" }] },
        },
      },
    ];
    Operators.getCSVs = jest.fn().mockReturnValue(csvs);

    const resources = [{ metadata: { name: "resource" } }];
    Operators.listResources = jest.fn().mockReturnValue({
      items: resources,
    });
    // Request resources for all namespaces
    await store.dispatch(operatorActions.getResources("default", ""));
    expect(Operators.listResources).toHaveBeenCalledWith(
      "default",
      "foo1-ns",
      "kubeapps.com/v1alpha1",
      "foo",
    );
    expect(Operators.listResources).toHaveBeenCalledWith(
      "default",
      "foo2-ns",
      "kubeapps.com/v1alpha1",
      "foo",
    );
  });

  it("ignores csv without owned crds", async () => {
    const csv = {
      metadata: { name: "foo" },
      spec: {
        customresourcedefinitions: {},
      },
    };
    Operators.getCSVs = jest.fn().mockReturnValue([csv]);
    Operators.listResources = jest.fn();
    const expectedActions = [
      {
        type: getType(operatorActions.requestCustomResources),
      },
      {
        type: getType(operatorActions.requestCSVs),
      },
      {
        type: getType(operatorActions.receiveCSVs),
        payload: [csv],
      },
      {
        type: getType(operatorActions.receiveCustomResources),
        payload: [],
      },
    ];
    await store.dispatch(operatorActions.getResources("default", "default"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(Operators.listResources).not.toHaveBeenCalled();
  });
});

describe("getResources", () => {
  it("get a resource in a namespace", async () => {
    const csv = {
      metadata: { name: "foo" },
      spec: {
        customresourcedefinitions: { owned: [{ name: "foo.kubeapps.com", version: "v1alpha1" }] },
      },
    };
    const resource = { metadata: { name: "resource" } };
    Operators.getCSV = jest.fn().mockReturnValue(csv);
    Operators.getResource = jest.fn().mockReturnValue(resource);
    const expectedActions = [
      {
        type: getType(operatorActions.requestCustomResource),
      },
      {
        type: getType(operatorActions.requestCSV),
      },
      {
        type: getType(operatorActions.receiveCSV),
        payload: csv,
      },
      {
        type: getType(operatorActions.receiveCustomResource),
        payload: resource,
      },
    ];
    await store.dispatch(
      operatorActions.getResource("default", "default", "foo", "foo.kubeapps.com", "bar"),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(Operators.getResource).toHaveBeenCalledWith(
      "default",
      "default",
      "kubeapps.com/v1alpha1",
      "foo",
      "bar",
    );
  });

  it("dispatches an error if getting a resource fails", async () => {
    const csv = {
      metadata: { name: "foo" },
      spec: {
        customresourcedefinitions: { owned: [{ name: "foo.kubeapps.com", version: "v1alpha1" }] },
      },
    };
    Operators.getCSV = jest.fn().mockReturnValue(csv);
    Operators.getResource = jest.fn(() => {
      throw new Error("Boom!");
    });
    const expectedActions = [
      {
        type: getType(operatorActions.requestCustomResource),
      },
      {
        type: getType(operatorActions.requestCSV),
      },
      {
        type: getType(operatorActions.receiveCSV),
        payload: csv,
      },
      {
        type: getType(operatorActions.errorCustomResource),
        payload: new Error("Boom!"),
      },
    ];
    await store.dispatch(
      operatorActions.getResource("default", "default", "foo", "foo.kubeapps.com", "bar"),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error if the given csv is not found", async () => {
    Operators.getCSV = jest.fn().mockReturnValue(undefined);
    const expectedActions = [
      {
        type: getType(operatorActions.requestCustomResource),
      },
      {
        type: getType(operatorActions.requestCSV),
      },
      {
        type: getType(operatorActions.receiveCSV),
      },
      {
        type: getType(operatorActions.errorCustomResource),
        payload: new Error("CSV foo not found in default"),
      },
    ];
    await store.dispatch(
      operatorActions.getResource("default", "default", "foo", "foo.kubeapps.com", "bar"),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error if the given crd is not found fails", async () => {
    const csv = {
      metadata: { name: "foo" },
      spec: {
        customresourcedefinitions: { owned: [{ name: "foo.kubeapps.com", version: "v1alpha1" }] },
      },
    };
    Operators.getCSV = jest.fn().mockReturnValue(csv);
    const expectedActions = [
      {
        type: getType(operatorActions.requestCustomResource),
      },
      {
        type: getType(operatorActions.requestCSV),
      },
      {
        type: getType(operatorActions.receiveCSV),
        payload: csv,
      },
      {
        type: getType(operatorActions.errorCustomResource),
        payload: new Error("Not found a valid CRD definition for foo/not-foo.kubeapps.com"),
      },
    ];
    await store.dispatch(
      operatorActions.getResource("default", "default", "foo", "not-foo.kubeapps.com", "bar"),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("deleteResource", () => {
  it("delete a resource in a namespace", async () => {
    const resource = { metadata: { name: "resource" } } as any;
    Operators.deleteResource = jest.fn();
    const expectedActions = [
      {
        type: getType(operatorActions.deletingResource),
      },
      {
        type: getType(operatorActions.resourceDeleted),
      },
    ];
    await store.dispatch(operatorActions.deleteResource("default", "default", "foos", resource));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error if deleting a resource fails", async () => {
    const resource = { metadata: { name: "resource" } } as any;
    Operators.deleteResource = jest.fn(() => {
      throw new Error("Boom!");
    });
    const expectedActions = [
      {
        type: getType(operatorActions.deletingResource),
      },
      {
        type: getType(operatorActions.errorResourceDelete),
        payload: new Error("Boom!"),
      },
    ];
    await store.dispatch(operatorActions.deleteResource("default", "default", "foos", resource));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("createOperator", () => {
  it("creates an Operator", async () => {
    const resource = {} as IResource;
    Operators.createOperator = jest.fn().mockReturnValue(resource);
    const expectedActions = [
      {
        type: getType(operatorActions.creatingOperator),
      },
      {
        type: getType(operatorActions.operatorCreated),
        payload: resource,
      },
    ];
    await store.dispatch(
      operatorActions.createOperator("default", "default", "etcd", "alpha", "Manual", "etcd.1.0.0"),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error", async () => {
    Operators.createOperator = jest.fn(() => {
      throw new Error("Boom!");
    });
    const expectedActions = [
      {
        type: getType(operatorActions.creatingOperator),
      },
      {
        type: getType(operatorActions.errorOperatorCreate),
        payload: new Error("Boom!"),
      },
    ];
    await store.dispatch(
      operatorActions.createOperator("default", "default", "etcd", "alpha", "Manual", "etcd.1.0.0"),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("listSubscriptions", () => {
  it("list subscriptions in the global namespace", async () => {
    const subscriptions = [{}] as IResource[];
    Operators.listSubscriptions = jest.fn().mockReturnValue({ items: subscriptions });
    const expectedActions = [
      {
        type: getType(operatorActions.requestSubscriptions),
      },
      {
        type: getType(operatorActions.receiveSubscriptions),
        payload: subscriptions,
      },
    ];
    await store.dispatch(operatorActions.listSubscriptions("default", "operators"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("list subscriptions in both the global and a single namespace", async () => {
    const subscriptions1 = [{ metadata: { name: "foo" } }] as IResource[];
    const subscriptions2 = [{ metadata: { name: "bar" } }] as IResource[];
    Operators.listSubscriptions = jest
      .fn()
      .mockReturnValueOnce({ items: subscriptions1 })
      .mockReturnValueOnce({ items: subscriptions2 });
    const expectedActions = [
      {
        type: getType(operatorActions.requestSubscriptions),
      },
      {
        type: getType(operatorActions.receiveSubscriptions),
        payload: subscriptions1.concat(subscriptions2),
      },
    ];
    await store.dispatch(operatorActions.listSubscriptions("default", "default"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error", async () => {
    Operators.listSubscriptions = jest.fn(() => {
      throw new Error("Boom!");
    });
    const expectedActions = [
      {
        type: getType(operatorActions.requestSubscriptions),
      },
      {
        type: getType(operatorActions.errorSubscriptionList),
        payload: new Error("Boom!"),
      },
    ];
    await store.dispatch(operatorActions.listSubscriptions("default", "default"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});
