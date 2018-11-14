import { shallow } from "enzyme";
import * as React from "react";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServicePlan } from "../../shared/ServiceCatalog";
import { IServiceInstance } from "../../shared/ServiceInstance";
import ServiceInstanceInfo from "./ServiceInstanceInfo";

const defaultName = "my-instance";
const defaultNS = "default";
const instance = {
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
} as IServiceInstance;

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
      displayName: defaultName,
    },
  },
} as IClusterServiceClass;

const plan = {
  metadata: {
    name: defaultName,
  },
  spec: {
    externalName: defaultName,
    externalID: `plan-${defaultName}`,
    description: "this is a plan",
    free: true,
  },
} as IServicePlan;

it("renders the Service Instance info", () => {
  const wrapper = shallow(
    <ServiceInstanceInfo instance={instance} svcClass={svcClass} plan={plan} />,
  );
  expect(wrapper.find(".ServiceInstanceInfo")).toExist();
  expect(wrapper).toMatchSnapshot();
});

it("renders a placeholder icon if service class has no icon", () => {
  const sClass = { ...svcClass, externalMetadata: {} };
  const wrapper = shallow(
    <ServiceInstanceInfo instance={instance} svcClass={sClass} plan={plan} />,
  );
  expect(wrapper.find(".ServiceInstanceInfo")).toExist();
  expect(wrapper).toMatchSnapshot();
});
