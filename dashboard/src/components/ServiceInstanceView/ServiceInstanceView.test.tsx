import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import { IServiceBindingWithSecret } from "shared/ServiceBinding";
import { IServiceInstance } from "shared/ServiceInstance";
import ServiceInstanceView from ".";
import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServicePlan } from "../../shared/ServiceCatalog";
import itBehavesLike from "../../shared/specs";
import { ForbiddenError, NotFoundError } from "../../shared/types";
import BindingListEntry from "../BindingList/BindingListEntry";
import Card from "../Card";
import { ErrorSelector } from "../ErrorAlert";
import AddBindingButton from "./AddBindingButton";

const defaultName = "my-instance";
const defaultNS = "default";
const defaultProps = {
  instances: { isFetching: false, list: [] },
  errors: {},
  bindingsWithSecrets: { isFetching: false, list: [] },
  name: defaultName,
  namespace: defaultNS,
  classes: { isFetching: false, list: [] },
  plans: { isFetching: false, list: [] },
  getInstances: jest.fn(),
  getClasses: jest.fn(),
  getPlans: jest.fn(),
  getBindings: jest.fn(),
  deprovision: jest.fn(),
  addBinding: jest.fn(),
  removeBinding: jest.fn(),
};

context("while fetching components", () => {
  const props = { ...defaultProps, classes: { isFetching: true, list: [] } };

  itBehavesLike("aLoadingComponent", { component: ServiceInstanceView, props });

  it("matches the snapshot", () => {
    const wrapper = shallow(<ServiceInstanceView {...props} />);
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a Application header", () => {
    const wrapper = shallow(<ServiceInstanceView {...props} />);
    expect(wrapper.find("h1").text()).toContain(defaultName);
  });
});

context("when all the components are loaded", () => {
  it("shows an error if the requested instance doesn't exist", () => {
    const wrapper = shallow(
      <ServiceInstanceView
        {...defaultProps}
        instances={{ isFetching: false, list: [{ metadata: { name: "" } } as IServiceInstance] }}
      />,
    );
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper.find(ErrorSelector).props()).toMatchObject({
      error: new NotFoundError(`Instance ${defaultName} not found in ${defaultNS}`),
    });
  });

  it("shows a fetch error if it exists", () => {
    const wrapper = shallow(
      <ServiceInstanceView {...defaultProps} errors={{ fetch: new ForbiddenError() }} />,
    );
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper).toMatchSnapshot();
  });

  it("shows a deprovision error if it exists", () => {
    const wrapper = shallow(
      <ServiceInstanceView {...defaultProps} errors={{ deprovision: new ForbiddenError() }} />,
    );
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper).toMatchSnapshot();
  });

  context("when an instance is available", () => {
    const instances = {
      isFetching: false,
      list: [
        {
          metadata: {
            name: defaultName,
            namespace: defaultNS,
            selfLink: "",
            uid: "",
            resourceVersion: "",
            creationTimestamp: "",
            finalizers: [],
            generation: 1,
          },
          spec: {
            clusterServiceClassExternalName: defaultName,
            clusterServiceClassRef: { name: defaultName },
            clusterServicePlanExternalName: defaultName,
            clusterServicePlanRef: { name: defaultName },
            externalID: defaultName,
          },
          status: {
            conditions: [
              {
                lastTransitionTime: "1",
                type: "a type",
                status: "good",
                reason: "none",
                message: "everything okay here",
              },
            ],
          },
        } as IServiceInstance,
      ],
    };

    it("should show the instance status info", () => {
      const wrapper = shallow(<ServiceInstanceView {...defaultProps} instances={instances} />);
      expect(wrapper.find(".found")).toExist();
      expect(wrapper).toMatchSnapshot();
    });

    it("should show class and plan info if it exists", () => {
      const classes = {
        isFetching: false,
        list: [
          {
            metadata: {
              name: defaultName,
              uid: `class-${defaultName}`,
            },
            spec: {
              bindable: true,
              externalName: defaultName,
              description: "this is a class",
              externalMetadata: {
                imageUrl: "img.png",
              },
            },
          } as IClusterServiceClass,
        ],
      };
      const plans = {
        isFetching: false,
        list: [
          {
            metadata: {
              name: defaultName,
            },
            spec: {
              externalName: defaultName,
              externalID: `plan-${defaultName}`,
              description: "this is a plan",
              free: true,
            },
          } as IServicePlan,
        ],
      };
      const wrapper = shallow(
        <ServiceInstanceView
          {...defaultProps}
          instances={instances}
          classes={classes}
          plans={plans}
        />,
      );
      expect(wrapper.find(Card).filterWhere(c => c.key() === `class-${defaultName}`)).toExist();
      expect(wrapper.find(Card).filterWhere(c => c.key() === `plan-${defaultName}`)).toExist();
      expect(wrapper).toMatchSnapshot();
    });

    it("should show the available bindings", () => {
      const classes = {
        isFetching: false,
        list: [
          {
            metadata: {
              name: defaultName,
              uid: `class-${defaultName}`,
            },
            spec: {
              bindable: true,
              externalName: defaultName,
              description: "this is a class",
              externalMetadata: {
                imageUrl: "img.png",
              },
            },
          } as IClusterServiceClass,
        ],
      };
      const bindings = {
        isFetching: false,
        list: [
          {
            binding: {
              metadata: {
                name: defaultName,
                namespace: defaultNS,
                uid: `binding-${defaultName}`,
              },
              spec: {
                instanceRef: {
                  name: defaultName,
                },
              },
              status: {
                conditions: [
                  {
                    message: "binding is okay",
                  },
                ],
              },
            },
          } as IServiceBindingWithSecret,
        ],
      };
      const wrapper = mount(
        <ServiceInstanceView
          {...defaultProps}
          instances={instances}
          classes={classes}
          bindingsWithSecrets={bindings}
        />,
      );
      expect(
        wrapper.find(BindingListEntry).filterWhere(b => b.key() === `binding-${defaultName}`),
      ).toExist();
      expect(wrapper).toMatchSnapshot();
    });

    it("should not show bindings information if the class is not bindable", () => {
      const classes = {
        isFetching: false,
        list: [
          {
            metadata: {
              name: defaultName,
              uid: `class-${defaultName}`,
            },
            spec: {
              bindable: false,
              externalName: defaultName,
              description: "this is a class",
              externalMetadata: {
                imageUrl: "img.png",
              },
            },
          } as IClusterServiceClass,
        ],
      };
      const wrapper = shallow(
        <ServiceInstanceView {...defaultProps} instances={instances} classes={classes} />,
      );
      expect(wrapper.find(AddBindingButton)).not.toExist();
      expect(wrapper.find(".found").text()).toContain("Instance Not Bindable");
    });
  });
});
