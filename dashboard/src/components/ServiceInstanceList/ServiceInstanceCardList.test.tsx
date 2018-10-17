import { shallow } from "enzyme";
import * as React from "react";
import ServiceInstanceCard from "./ServiceInstanceCard";
import ServiceInstanceCardList from "./ServiceInstanceCardList";

it("parses a list of instances and classes", () => {
  const wrapper = shallow(
    <ServiceInstanceCardList
      instances={[
        {
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
        },
      ]}
      classes={[
        {
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
        },
      ]}
    />,
  );
  expect(wrapper.find(ServiceInstanceCard).props()).toMatchObject({
    link: "/services/brokers/foo/instances/ns/foobar/foo",
    name: "foo",
    namespace: "foobar",
    servicePlanName: "foo-plan",
    statusReason: "Provisioned",
  });
});
