import { CdsButton } from "components/Clarity/clarity";
import ConfirmDialog from "components/ConfirmDialog/ConfirmDialog.v2";
import AdvancedDeploymentForm from "components/DeploymentFormBody/AdvancedDeploymentForm.v2";
import Alert from "components/js/Alert";
import { mount } from "enzyme";
import * as React from "react";
import { act } from "react-dom/test-utils";
import itBehavesLike from "../../shared/specs";
import OperatorInstanceFormBody from "./OperatorInstanceFormBody.v2";
import { IOperatorInstanceFormProps } from "./OperatorInstanceFormBody.v2";

const defaultProps: IOperatorInstanceFormProps = {
  isFetching: false,
  namespace: "kubeapps",
  handleDeploy: jest.fn(),
  defaultValues: "",
};

itBehavesLike("aLoadingComponent", {
  component: OperatorInstanceFormBody,
  props: { ...defaultProps, isFetching: true },
});

it("set default values", () => {
  const wrapper = mount(<OperatorInstanceFormBody {...defaultProps} defaultValues="foo" />);
  expect(wrapper.find(AdvancedDeploymentForm).prop("appValues")).toBe("foo");
});

it("renders an error if the namespace is _all", () => {
  const wrapper = mount(<OperatorInstanceFormBody {...defaultProps} namespace="_all" />);
  expect(wrapper.find(Alert)).toIncludeText("Select a namespace before creating a new instance");
});

it("restores the default values", async () => {
  const wrapper = mount(<OperatorInstanceFormBody {...defaultProps} defaultValues="foo" />);

  act(() => {
    (wrapper.find(AdvancedDeploymentForm).prop("handleValuesChange") as any)("not-foo");
  });
  wrapper.update();
  expect(wrapper.find(AdvancedDeploymentForm).prop("appValues")).toBe("not-foo");

  const restoreButton = wrapper
    .find(CdsButton)
    .filterWhere(b => b.text().includes("Restore Defaults"));
  act(() => {
    restoreButton.simulate("click");
  });
  act(() => {
    (wrapper.find(ConfirmDialog).prop("onConfirm") as any)();
  });
  wrapper.update();

  expect(wrapper.find(AdvancedDeploymentForm).prop("appValues")).toBe("foo");
});

it("should submit the form", () => {
  const handleDeploy = jest.fn();
  const wrapper = mount(<OperatorInstanceFormBody {...defaultProps} handleDeploy={handleDeploy} />);

  const values = "apiVersion: v1\nmetadata:\n  name: foo";
  act(() => {
    (wrapper.find(AdvancedDeploymentForm).prop("handleValuesChange") as any)(values);
  });
  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  const resource = {
    apiVersion: "v1",
    metadata: {
      name: "foo",
    },
  };
  expect(handleDeploy).toHaveBeenCalledWith(resource);
});

it("should catch a syntax error in the form", () => {
  const handleDeploy = jest.fn();
  const wrapper = mount(<OperatorInstanceFormBody {...defaultProps} handleDeploy={handleDeploy} />);

  const values = "metadata: invalid!\n  name: foo";
  act(() => {
    (wrapper.find(AdvancedDeploymentForm).prop("handleValuesChange") as any)(values);
  });
  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  expect(wrapper.find(Alert)).toIncludeText("Unable to parse the given YAML. Got: bad indentation");
  expect(handleDeploy).not.toHaveBeenCalled();
});

it("should throw an eror if the element doesn't contain an apiVersion", () => {
  const handleDeploy = jest.fn();
  const wrapper = mount(<OperatorInstanceFormBody {...defaultProps} handleDeploy={handleDeploy} />);

  const values = "metadata:\nname: foo";
  act(() => {
    (wrapper.find(AdvancedDeploymentForm).prop("handleValuesChange") as any)(values);
  });
  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  expect(wrapper.find(Alert)).toIncludeText(
    "Unable parse the resource. Make sure it contains a valid apiVersion",
  );
  expect(handleDeploy).not.toHaveBeenCalled();
});
