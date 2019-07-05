import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import ServiceClassView from ".";
import { ErrorSelector } from "../../components/ErrorAlert";
import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServicePlan } from "../../shared/ServiceCatalog";
import itBehavesLike from "../../shared/specs";
import { ForbiddenError } from "../../shared/types";
import ServiceClassPlan from "./ServiceClassPlan";

const defaultName = "my-class";
const defaultNS = "default";
const defaultProps = {
  classes: { isFetching: false, list: [] },
  classname: defaultName,
  getClasses: jest.fn(),
  getPlans: jest.fn(),
  plans: { isFetching: false, list: [] },
  provision: jest.fn(),
  push: jest.fn(),
  namespace: defaultNS,
};

context("while fetching components", () => {
  const props = { ...defaultProps, classes: { isFetching: true, list: [] } };

  itBehavesLike("aLoadingComponent", { component: ServiceClassView, props });

  it("matches the snapshot", () => {
    const wrapper = shallow(<ServiceClassView {...props} />);
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a Application header", () => {
    const wrapper = shallow(<ServiceClassView {...props} />);
    expect(wrapper.find("h1").text()).toContain(defaultName);
  });
});

context("when all the components are loaded", () => {
  it("shows an error when listing classes if it exists", () => {
    const wrapper = shallow(<ServiceClassView {...defaultProps} error={new ForbiddenError()} />);
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper.find(ErrorSelector).props()).toMatchObject({
      error: new ForbiddenError(),
    });
  });

  context("when an service class is available", () => {
    let classes: any = {};
    let plans: any = {};
    beforeEach(() => {
      classes = {
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
      plans = {
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
              clusterServiceClassRef: {
                name: defaultName,
              },
              free: true,
            },
          } as IServicePlan,
        ],
      };
    });

    it("should show the class info", () => {
      const wrapper = shallow(
        <ServiceClassView {...defaultProps} classes={classes} plans={plans} />,
      );
      expect(wrapper.find(ServiceClassPlan)).toExist();
      expect(wrapper).toMatchSnapshot();
    });

    it("should avoid plans that doesn't belong to the class", () => {
      const newPlan = {
        metadata: {
          name: defaultName,
        },
        spec: {
          externalName: defaultName,
          externalID: `plan-${defaultName}`,
          description: "this is a plan",
          clusterServiceClassRef: {
            name: `not-${defaultName}`,
          },
          free: true,
        },
      } as IServicePlan;
      plans.list.push(newPlan);
      const wrapper = shallow(
        <ServiceClassView {...defaultProps} classes={classes} plans={plans} />,
      );
      const planList = wrapper.find(ServiceClassPlan);
      expect(planList).toExist();
      expect(planList.length).toBe(1);
    });
  });
});
