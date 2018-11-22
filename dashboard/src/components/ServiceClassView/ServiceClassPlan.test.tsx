import { shallow } from "enzyme";
import * as React from "react";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServicePlan } from "../../shared/ServiceCatalog";
import ServiceClassPlan from "./ServiceClassPlan";

const defaultName = "my-class";
const defaultNS = "default";
const svcClass = {
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
} as IClusterServiceClass;
const svcPlan = {
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
} as IServicePlan;
const defaultProps = {
  svcClass,
  svcPlan,
  provision: jest.fn(),
  push: jest.fn(),
  namespace: defaultNS,
};

it("renders a service plan", () => {
  const wrapper = shallow(<ServiceClassPlan {...defaultProps} />);
  expect(wrapper).toMatchSnapshot();
});
