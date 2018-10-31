import { shallow } from "enzyme";
import * as React from "react";

import { IServiceBindingWithSecret } from "shared/ServiceBinding";
import BindingList from "./BindingList";

it("renders information about the binding with a secret", () => {
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
  const wrapper = shallow(
    <BindingList bindingsWithSecrets={[bindingWithSecret]} removeBinding={jest.fn()} />,
  );
  expect(wrapper).toMatchSnapshot();
});
