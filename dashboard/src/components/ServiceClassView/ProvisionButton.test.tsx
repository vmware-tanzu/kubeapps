import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import { JSONSchema6 } from "json-schema";
import { ErrorSelector } from "../../components/ErrorAlert";
import SchemaForm from "../../components/SchemaForm";
import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServicePlan } from "../../shared/ServiceCatalog";
import ProvisionButton from "./ProvisionButton";

const defaultName = "my-class";
const defaultNS = "default";
const selectedClass = {
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
let selectedPlan: any = {};
let defaultProps = {
  selectedClass,
  selectedPlan,
  provision: jest.fn(),
  push: jest.fn(),
  namespace: defaultNS,
};

beforeEach(() => {
  selectedPlan = {
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
  defaultProps = {
    selectedClass,
    selectedPlan,
    provision: jest.fn(),
    push: jest.fn(),
    namespace: defaultNS,
  };
});

it("shows a provisioning message", () => {
  const wrapper = shallow(<ProvisionButton {...defaultProps} />);
  wrapper.setState({ isProvisioning: true });
  expect(wrapper.text()).toContain("Provisioning...");
});

context("when the modal is open", () => {
  it("shows an error if is present", () => {
    const wrapper = shallow(<ProvisionButton {...defaultProps} error={new Error()} />);
    wrapper.setState({ modalIsOpen: true });

    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper).toMatchSnapshot();
  });

  it("should display the name form", () => {
    const wrapper = shallow(<ProvisionButton {...defaultProps} />);
    wrapper.setState({ modalIsOpen: true, displayNameForm: true });

    const form = wrapper.find(SchemaForm);
    expect(form).toExist();
    expect((form.prop("schema") as any).properties).toHaveProperty("Name");
  });

  it("should display the default schema if it's not in the plan spec", () => {
    selectedPlan.spec.instanceCreateParameterSchema = null;
    const wrapper = shallow(<ProvisionButton {...defaultProps} />);
    wrapper.setState({ modalIsOpen: true, displayNameForm: false });

    const form = wrapper.find(SchemaForm);
    expect(form).toExist();
    const schema = form.prop("schema") as any;
    expect(schema).toMatchObject({
      properties: {
        kubeappsRawParameters: {
          title: "Parameters",
          type: "object",
        },
      },
      type: "object",
    });
  });

  it("should display the spec schema", () => {
    const newSchema = {
      properties: {
        kubeappsRawParameters: {
          title: "Foo",
          type: "string",
        },
      },
      type: "object",
    } as JSONSchema6;
    selectedPlan.spec.instanceCreateParameterSchema = newSchema;
    const wrapper = shallow(<ProvisionButton {...defaultProps} />);
    wrapper.setState({ modalIsOpen: true, displayNameForm: false });

    const form = wrapper.find(SchemaForm);
    expect(form).toExist();
    const schema = form.prop("schema") as any;
    expect(schema).toEqual(newSchema);
  });
});
