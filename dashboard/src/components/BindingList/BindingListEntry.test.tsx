import { shallow } from "enzyme";
import * as React from "react";

import { IServiceBindingWithSecret } from "shared/ServiceBinding";
import BindingListEntry from "./BindingListEntry";

const bindingWithSecret = {
  binding: {
    metadata: {
      name: "foo",
      namespace: "default",
      selfLink: "",
      uid: "",
      resourceVersion: "",
      creationTimestamp: "",
      finalizers: [],
      generation: 1,
    },
    spec: {
      externalID: "",
      instanceRef: {
        name: "foo",
      },
      secretName: "bar",
    },
    status: {
      asyncOpInProgress: false,
      currentOperation: "",
      reconciledGeneration: 1,
      operationStartTime: "",
      externalProperties: {},
      orphanMitigationInProgress: false,
      unbindStatus: "",
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
  },
  secret: {
    kind: "",
    apiVersion: "",
    metadata: {
      name: "bar",
      namespace: "default",
      selfLink: "",
      uid: "",
      resourceVersion: "",
      creationTimestamp: "",
      finalizers: [],
      generation: 1,
    },
    data: {
      mySecret: "Y29udGVudAo=",
    },
  },
} as IServiceBindingWithSecret;

it("renders information about the binding with a secret", () => {
  const wrapper = shallow(
    <BindingListEntry bindingWithSecret={bindingWithSecret} removeBinding={jest.fn()} />,
  );
  expect(wrapper).toMatchSnapshot();
});

it("if there are several conditions, shows the latest one", () => {
  const bindingWithConditions = { ...bindingWithSecret };
  bindingWithConditions.binding.status.conditions = [
    {
      lastTransitionTime: "2018-10-17T09:47:24Z",
      type: "a type",
      status: "good",
      reason: "notAReason",
      message: "everything okay here",
    },
    {
      lastTransitionTime: "2018-10-17T09:47:34Z",
      type: "a type",
      status: "notGood",
      reason: "aGoodReason",
      message: "failure!",
    },
  ];
  const wrapper = shallow(
    <BindingListEntry bindingWithSecret={bindingWithSecret} removeBinding={jest.fn()} />,
  );
  expect(wrapper.text()).toContain("aGoodReason");
  expect(wrapper.text()).not.toContain("notAReason");
});
