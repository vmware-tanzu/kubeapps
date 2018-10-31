import { shallow } from "enzyme";
import * as React from "react";

import { IServiceBindingWithSecret } from "shared/ServiceBinding";
import TerminalModal from "../../components/TerminalModal";
import BindingDetails from "./BindingDetails";

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
  const wrapper = shallow(<BindingDetails {...bindingWithSecret} />);
  expect(wrapper.text()).toContain("bar (show)");
  expect(wrapper).toMatchSnapshot();
});

it("should show the secret content when clicking the button", () => {
  const wrapper = shallow(<BindingDetails {...bindingWithSecret} />);
  const button = wrapper.find("a");
  expect(button).toExist();

  expect((wrapper.state() as any).modalIsOpen).toBe(false);
  button.simulate("click");
  expect((wrapper.state() as any).modalIsOpen).toBe(true);
  expect(wrapper.find(TerminalModal).props().message).toContain("mySecret: content");
});
