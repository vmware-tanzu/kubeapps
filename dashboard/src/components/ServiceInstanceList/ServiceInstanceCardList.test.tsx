import { shallow } from "enzyme";
import * as React from "react";
import ServiceInstanceCard from "./ServiceInstanceCard";
import ServiceInstanceCardList from "./ServiceInstanceCardList";

const instance = {
  metadata: {
    name: "foo",
    namespace: "foobar",
    selfLink: "",
    uid: "",
    finalizers: [],
    resourceVersion: "",
    creationTimestamp: "",
    generation: 1,
  },
  spec: {
    clusterServiceClassRef: {
      name: "foo-class",
    },
    clusterServiceClassExternalName: "foo-class",
    clusterServicePlanExternalName: "foo-plan",
    externalID: "",
  },
  status: {
    conditions: [
      {
        status: "",
        lastTransitionTime: "",
        message: "",
        type: "",
        reason: "Provisioned",
      },
    ],
  },
};

const svcClass = {
  metadata: {
    name: "foo-class",
    selfLink: "",
    uid: "",
    resourceVersion: "",
    creationTimestamp: "",
  },
  spec: {
    bindable: true,
    binding_retrievable: true,
    clusterServiceBrokerName: "foo",
    description: "",
    externalID: "",
    externalName: "",
    planUpdatable: true,
    tags: [],
  },
  status: {
    removedFromBrokerCatalog: false,
  },
};

it("parses a list of instances and classes", () => {
  const wrapper = shallow(<ServiceInstanceCardList instances={[instance]} classes={[svcClass]} />);
  expect(wrapper.find(ServiceInstanceCard).props()).toMatchObject({
    link: "/services/brokers/foo/instances/ns/foobar/foo",
    name: "foo",
    namespace: "foobar",
    servicePlanName: "foo-plan",
    statusReason: "provisioned",
  });
});

it("Split into words the status of the service instance", () => {
  const newInstance = Object.assign({}, instance);
  newInstance.status.conditions[0].reason = "ClusterServiceBrokerReturnedFailure";
  const wrapper = shallow(<ServiceInstanceCardList instances={[instance]} classes={[svcClass]} />);
  expect(wrapper.find(ServiceInstanceCard).props()).toMatchObject({
    link: "/services/brokers/foo/instances/ns/foobar/foo",
    name: "foo",
    namespace: "foobar",
    servicePlanName: "foo-plan",
    statusReason: "clusterServiceBrokerReturnedFailure",
  });
});
