import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import { IServiceBindingWithSecret } from "shared/ServiceBinding";
import { IServiceInstance } from "shared/ServiceInstance";
import ServiceInstanceView from ".";
import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import itBehavesLike from "../../shared/specs";
import { ForbiddenError, NotFoundError } from "../../shared/types";
import BindingListEntry from "../BindingList/BindingListEntry";
import { ErrorSelector } from "../ErrorAlert";
import AddBindingButton from "./AddBindingButton";
import DeprovisionButton from "./DeprovisionButton";
import ServiceInstanceInfo from "./ServiceInstanceInfo";
import ServiceInstanceStatus from "./ServiceInstanceStatus";

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
    expect(wrapper.find(ServiceInstanceInfo)).not.toExist();
  });

  it("shows a fetch error if it exists", () => {
    const wrapper = shallow(
      <ServiceInstanceView {...defaultProps} errors={{ fetch: new ForbiddenError() }} />,
    );
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper).toMatchSnapshot();
    expect(wrapper.find(ServiceInstanceInfo)).not.toExist();
  });

  it("shows a deprovision error if it exists", () => {
    const wrapper = shallow(
      <ServiceInstanceView {...defaultProps} errors={{ deprovision: new ForbiddenError() }} />,
    );
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper).toMatchSnapshot();
    expect(wrapper.find(ServiceInstanceInfo)).not.toExist();
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

    it("should show the instance status info", () => {
      const wrapper = shallow(<ServiceInstanceView {...defaultProps} instances={instances} />);
      expect(wrapper.find(ServiceInstanceInfo)).toExist();
      expect(wrapper.find(ServiceInstanceStatus)).toExist();
      expect(wrapper.find(".ServiceInstanceView__details")).toExist();
      expect(wrapper).toMatchSnapshot();
    });

    it("should show the available bindings", () => {
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
      const classesNotBindable = {
        ...classes,
        list: [{ ...classes.list[0], spec: { ...classes.list[0].spec, bindable: false } }],
      };
      const wrapper = shallow(
        <ServiceInstanceView
          {...defaultProps}
          instances={instances}
          classes={classesNotBindable}
        />,
      );
      expect(wrapper.find(AddBindingButton)).not.toExist();
      expect(wrapper.find(".ServiceInstanceView__details").text()).toContain(
        "This instance cannot be bound to applications",
      );
    });

    const instancesWithReason = (reason: string) => ({
      ...instances,
      list: [
        {
          ...instances.list[0],
          status: {
            conditions: [
              {
                ...instances.list[0].status.conditions[0],
                reason,
              },
            ],
          },
        },
      ],
    });

    const buttonStateTestCases = [
      { reason: "Provisioned", disabled: false },
      { reason: "Unknown", disabled: false },
      { reason: "Failed", disabled: false },
      { reason: "Provisioning", disabled: true },
      { reason: "Deprovisioning", disabled: true },
    ];
    buttonStateTestCases.forEach(t => {
      it(`should have the correct button state when in the ${t.reason} state`, () => {
        const is = instancesWithReason(t.reason);
        const wrapper = shallow(
          <ServiceInstanceView {...defaultProps} instances={is} classes={classes} />,
        );
        const deprovisionButton = wrapper.find(DeprovisionButton);
        const addBindingButton = wrapper.find(AddBindingButton);
        expect(deprovisionButton).toExist();
        expect(addBindingButton).toExist();
        expect(deprovisionButton.props().disabled).toBe(t.disabled);
      });
    });
  });
});
