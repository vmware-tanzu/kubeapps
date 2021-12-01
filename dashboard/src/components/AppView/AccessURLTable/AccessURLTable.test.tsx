import actions from "actions";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import context from "jest-plugin-context";
import * as ReactRedux from "react-redux";
import ResourceRef, { keyForResourceRef } from "shared/ResourceRef";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IIngressSpec, IResource, IServiceSpec, IServiceStatus } from "shared/types";
import AccessURLTable from "./AccessURLTable";

const defaultProps = {
  serviceRefs: [],
  ingressRefs: [],
};

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.kube = {
    ...actions.kube,
    getResource: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

describe("when receiving ingresses", () => {
  it("fetches ingresses at mount time", () => {
    const ingress = { name: "ing", getResourceURL: jest.fn() } as any;
    const mock = jest.fn();
    actions.kube.getResource = mock;
    mountWrapper(defaultStore, <AccessURLTable {...defaultProps} ingressRefs={[ingress]} />);
    expect(mock).toHaveBeenCalledWith(ingress);
  });

  it("fetches when new ingress refs received", () => {
    const ingress = { name: "ing", getResourceURL: jest.fn() } as any;
    const mock = jest.fn();
    actions.kube.getResource = mock;
    const wrapper = mountWrapper(
      defaultStore,
      <AccessURLTable {...defaultProps} ingressRefs={[ingress]} />,
    );
    wrapper.setProps({ ingressRefs: [ingress] });
    expect(mock).toHaveBeenCalledWith(ingress);
  });
});

context("when some resource is fetching", () => {
  it("shows a loadingWrapper when fetching services", () => {
    const serviceItem = { isFetching: true };
    const svcRef = {
      apiVersion: "v1",
      kind: "Service",
      namespace: "test",
      name: "svc",
    } as ResourceRef;
    const serviceRefs = [svcRef];
    const svcKey = keyForResourceRef(svcRef.apiVersion, svcRef.kind, svcRef.namespace, svcRef.name);
    const state = {
      kube: { items: { [svcKey]: serviceItem } },
    };
    const store = getStore(state);
    const wrapper = mountWrapper(
      store,
      <AccessURLTable {...defaultProps} serviceRefs={serviceRefs} />,
    );

    expect(wrapper.find(LoadingWrapper)).toExist();
  });

  it("displays the error (while fetching)", () => {
    const ingressItem = { isFetching: true };
    const ingressRef = {
      apiVersion: "v1",
      kind: "Ingress",
      namespace: "test",
      name: "ingress",
    } as ResourceRef;
    const ingressRefs = [ingressRef];
    const ingressKey = keyForResourceRef(
      ingressRef.apiVersion,
      ingressRef.kind,
      ingressRef.namespace,
      ingressRef.name,
    );
    const state = {
      kube: { items: { [ingressKey]: ingressItem } },
    };
    const store = getStore(state);
    const wrapper = mountWrapper(
      store,
      <AccessURLTable {...defaultProps} ingressRefs={ingressRefs} />,
    );

    expect(wrapper.find(LoadingWrapper)).toExist();
  });
});

it("doesn't render anything if the application has no URL", () => {
  const wrapper = mountWrapper(defaultStore, <AccessURLTable {...defaultProps} />);
  expect(wrapper.find("table")).not.toExist();
});

context("when the app contains services", () => {
  it("should omit the Service Table if there are no public services", () => {
    const service = {
      kind: "Service",
      metadata: {
        name: "foo",
        selfLink: "/services/foo",
      },
      spec: {
        type: "ClusterIP",
        ports: [{ port: 8080 }],
      } as IServiceSpec,
      status: {
        loadBalancer: {},
      } as IServiceStatus,
    } as IResource;
    const serviceItem = { isFetching: false, item: service };
    const svcRef = {
      apiVersion: "v1",
      kind: "Service",
      namespace: "test",
      name: "svc",
    } as ResourceRef;
    const serviceRefs = [svcRef];
    const svcKey = keyForResourceRef(svcRef.apiVersion, svcRef.kind, svcRef.namespace, svcRef.name);
    const state = { kube: { items: { [svcKey]: serviceItem } } };
    const store = getStore(state);
    const wrapper = mountWrapper(
      store,
      <AccessURLTable {...defaultProps} serviceRefs={serviceRefs} />,
    );
    expect(wrapper.text()).toContain("The current application does not expose a public URL");
  });

  it("should show the table if any service is a LoadBalancer", () => {
    const service = {
      kind: "Service",
      metadata: {
        name: "foo",
        selfLink: "/services/foo",
      },
      spec: {
        type: "LoadBalancer",
        ports: [{ port: 8080 }],
      } as IServiceSpec,
      status: {
        loadBalancer: {},
      } as IServiceStatus,
    } as IResource;
    const serviceItem = { isFetching: false, item: service };
    const svcRef = {
      apiVersion: "v1",
      kind: "Service",
      namespace: "test",
      name: "svc",
    } as ResourceRef;
    const serviceRefs = [svcRef];
    const svcKey = keyForResourceRef(svcRef.apiVersion, svcRef.kind, svcRef.namespace, svcRef.name);
    const state = { kube: { items: { [svcKey]: serviceItem } } };
    const store = getStore(state);
    const wrapper = mountWrapper(
      store,
      <AccessURLTable {...defaultProps} serviceRefs={serviceRefs} />,
    );
    expect(wrapper.find("Table")).toExist();
  });
});

context("when the app contains ingresses", () => {
  it("should show the table with available ingresses", () => {
    const ingress = {
      kind: "Ingress",
      metadata: {
        name: "foo",
        selfLink: "/ingresses/foo",
      },
      spec: {
        rules: [
          {
            host: "foo.bar",
            http: {
              paths: [{ path: "/ready" }],
            },
          },
        ],
      } as IIngressSpec,
    } as IResource;
    const ingressItem = { isFetching: false, item: ingress };
    const ingressRef = {
      apiVersion: "v1",
      kind: "Ingress",
      namespace: "test",
      name: "ingress",
    } as ResourceRef;
    const ingressRefs = [ingressRef];
    const ingressKey = keyForResourceRef(
      ingressRef.apiVersion,
      ingressRef.kind,
      ingressRef.namespace,
      ingressRef.name,
    );
    const state = { kube: { items: { [ingressKey]: ingressItem } } };
    const store = getStore(state);
    const wrapper = mountWrapper(
      store,
      <AccessURLTable {...defaultProps} ingressRefs={ingressRefs} />,
    );
    expect(wrapper.find("Table")).toExist();
    expect(wrapper.find("a").findWhere(a => a.prop("href") === "http://foo.bar/ready")).toExist();
  });

  it("should show the table with available ingresses without anchors if a regex is present in the path", () => {
    const ingress = {
      kind: "Ingress",
      metadata: {
        name: "foo",
        selfLink: "/ingresses/foo",
      },
      spec: {
        rules: [
          {
            host: "foo.bar",
            http: {
              paths: [{ path: "/ready(/|$)(.*)" }],
            },
          },
        ],
      } as IIngressSpec,
    } as IResource;
    const ingressItem = { isFetching: false, item: ingress };
    const ingressRef = {
      apiVersion: "v1",
      kind: "Ingress",
      namespace: "test",
      name: "ingress",
    } as ResourceRef;
    const ingressRefs = [ingressRef];
    const ingressKey = keyForResourceRef(
      ingressRef.apiVersion,
      ingressRef.kind,
      ingressRef.namespace,
      ingressRef.name,
    );
    const state = { kube: { items: { [ingressKey]: ingressItem } } };
    const store = getStore(state);
    const wrapper = mountWrapper(
      store,
      <AccessURLTable {...defaultProps} ingressRefs={ingressRefs} />,
    );
    expect(wrapper.find("Table")).toExist();
    expect(wrapper.find("a")).not.toExist();
    const matchingSpans = wrapper.find("span").findWhere(s => s.text().includes("foo.bar/ready"));
    expect(matchingSpans).not.toHaveLength(0);
    matchingSpans.forEach(element => {
      expect(element.text()).toEqual("http://foo.bar/ready(/|$)(.*)");
    });
  });
});

context("when the app contains services and ingresses", () => {
  it("should show the table with available svcs and ingresses", () => {
    const service = {
      kind: "Service",
      metadata: {
        name: "foo",
        selfLink: "/services/foo",
      },
      spec: {
        type: "LoadBalancer",
        ports: [{ port: 8080 }],
      } as IServiceSpec,
      status: {
        loadBalancer: {
          ingress: [
            {
              ip: "1.2.3.4",
            },
          ],
        },
      } as IServiceStatus,
    } as IResource;
    const serviceItem = { isFetching: false, item: service };
    const svcRef = {
      apiVersion: "v1",
      kind: "Service",
      namespace: "test",
      name: "svc",
    } as ResourceRef;
    const serviceRefs = [svcRef];
    const ingress = {
      kind: "Ingress",
      metadata: {
        name: "foo",
        selfLink: "/ingresses/foo",
      },
      spec: {
        rules: [
          {
            host: "foo.bar",
            http: {
              paths: [{ path: "/ready" }],
            },
          },
        ],
      } as IIngressSpec,
    } as IResource;
    const ingressItem = { isFetching: false, item: ingress };
    const ingressRef = {
      apiVersion: "v1",
      kind: "Ingress",
      namespace: "test",
      name: "ingress",
    } as ResourceRef;
    const ingressRefs = [ingressRef];
    const svcKey = keyForResourceRef(svcRef.apiVersion, svcRef.kind, svcRef.namespace, svcRef.name);
    const ingressKey = keyForResourceRef(
      ingressRef.apiVersion,
      ingressRef.kind,
      ingressRef.namespace,
      ingressRef.name,
    );
    const state = {
      kube: { items: { [svcKey]: serviceItem, [ingressKey]: ingressItem } },
    };
    const store = getStore(state);
    const wrapper = mountWrapper(
      store,
      <AccessURLTable {...defaultProps} ingressRefs={ingressRefs} serviceRefs={serviceRefs} />,
    );
    expect(wrapper.find("a").findWhere(a => a.prop("href") === "http://1.2.3.4:8080")).toExist();
    expect(wrapper.find("a").findWhere(a => a.prop("href") === "http://foo.bar/ready")).toExist();
  });
});

context("when the app contains resources with errors", () => {
  it("displays the error (when resources with errors)", () => {
    const serviceItem = { isFetching: false, error: new Error("could not find Service") };
    const svcRef = {
      apiVersion: "v1",
      kind: "Service",
      namespace: "test",
      name: "svc",
    } as ResourceRef;
    const serviceRefs = [svcRef];
    const ingressItem = { isFetching: false, error: new Error("could not find Ingress") };
    const ingressRef = {
      apiVersion: "v1",
      kind: "Ingress",
      namespace: "test",
      name: "ingress",
    } as ResourceRef;
    const ingressRefs = [ingressRef];
    const svcKey = keyForResourceRef(svcRef.apiVersion, svcRef.kind, svcRef.namespace, svcRef.name);
    const ingressKey = keyForResourceRef(
      ingressRef.apiVersion,
      ingressRef.kind,
      ingressRef.namespace,
      ingressRef.name,
    );
    const state = {
      kube: { items: { [svcKey]: serviceItem, [ingressKey]: ingressItem } },
    };
    const store = getStore(state);
    const wrapper = mountWrapper(
      store,
      <AccessURLTable {...defaultProps} serviceRefs={serviceRefs} ingressRefs={ingressRefs} />,
    );

    // The Service error is not shown, as it is filtered out because without the
    // resource we can't determine whether it is a public LoadBalancer Service
    // or not. The Service error will be shown in the Services table anyway.
    expect(wrapper.find("a")).not.toExist();
    expect(wrapper.find("table").text()).toContain("could not find Ingress");
  });
});
