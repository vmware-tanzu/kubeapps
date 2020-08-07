import { IPartialAppViewState } from "components/AppView/AppView";
import OperatorNotSupported from "components/OperatorList/OperatorsNotSupported";
import { shallow } from "enzyme";
import * as React from "react";
import Modal from "react-modal";
import AccessURLTable from "../../containers/AccessURLTableContainer";
import ApplicationStatus from "../../containers/ApplicationStatusContainer";
import itBehavesLike from "../../shared/specs";
import { app } from "../../shared/url";
import AppNotes from "../AppView/AppNotes";
import AppValues from "../AppView/AppValues";
import ResourceTable from "../AppView/ResourceTable";
import ConfirmDialog from "../ConfirmDialog";
import { ErrorSelector } from "../ErrorAlert";
import OperatorInstance, { IOperatorInstanceProps } from "./OperatorInstance";

const defaultProps: IOperatorInstanceProps = {
  isFetching: false,
  cluster: "default",
  namespace: "default",
  csvName: "foo",
  crdName: "foo.kubeapps.com",
  instanceName: "foo-cluster",
  getResource: jest.fn(),
  deleteResource: jest.fn(),
  push: jest.fn(),
  errors: {},
};

itBehavesLike("aLoadingComponent", {
  component: OperatorInstance,
  props: { ...defaultProps, isFetching: true },
});

it("displays an alert if rendered for an additional cluster", () => {
  const props = { ...defaultProps, cluster: "other-cluster" };
  const wrapper = shallow(<OperatorInstance {...props} />);
  expect(wrapper.find(OperatorNotSupported)).toExist();
});

it("gets a resource when loading the component", () => {
  const getResource = jest.fn();
  shallow(<OperatorInstance {...defaultProps} getResource={getResource} />);
  expect(getResource).toHaveBeenCalledWith(
    defaultProps.namespace,
    defaultProps.csvName,
    defaultProps.crdName,
    defaultProps.instanceName,
  );
});

it("gets a resource again if the namespace changes", () => {
  const getResource = jest.fn();
  const wrapper = shallow(<OperatorInstance {...defaultProps} getResource={getResource} />);
  wrapper.setProps({ namespace: "other" });
  expect(getResource).toHaveBeenCalledTimes(2);
});

it("renders an error", () => {
  const wrapper = shallow(
    <OperatorInstance {...defaultProps} errors={{ fetch: new Error("Boom!") }} />,
  );
  expect(wrapper.find(ErrorSelector)).toExist();
  expect(wrapper.find(AppNotes)).not.toExist();
});

describe("renders a resource", () => {
  const csv = {
    metadata: { name: "foo" },
    spec: {
      icon: [{}],
      customresourcedefinitions: {
        owned: [
          {
            name: "foo.kubeapps.com",
            version: "v1alpha1",
            kind: "Foo",
            resources: [{ kind: "Deployment" }],
          },
        ],
      },
    },
  } as any;
  const resource = {
    kind: "Foo",
    metadata: { name: "foo-instance" },
    spec: { test: true },
    status: { alive: true },
  } as any;

  it("renders the resource and CSV info", () => {
    const wrapper = shallow(<OperatorInstance {...defaultProps} />);
    wrapper.setProps({ csv, resource });
    expect(wrapper.find(AppNotes)).toExist();
    expect(wrapper.find(AppValues)).toExist();
    expect(wrapper.find(".ChartInfo")).toExist();
    expect(wrapper.find(ApplicationStatus)).toExist();
    expect(wrapper.find(AccessURLTable)).toExist();
    expect(wrapper.find(ResourceTable)).toExist();
    expect(wrapper).toMatchSnapshot();
  });

  it("deletes the resource", async () => {
    const deleteResource = jest.fn().mockReturnValue(true);
    const push = jest.fn();
    const wrapper = shallow(
      <OperatorInstance {...defaultProps} deleteResource={deleteResource} push={push} />,
    );
    wrapper.setProps({ csv, resource });
    Modal.setAppElement(document.createElement("div"));
    wrapper.find(".button-danger").simulate("click");

    const dialog = wrapper.find(ConfirmDialog);
    expect(dialog.prop("modalIsOpen")).toEqual(true);
    (dialog.prop("onConfirm") as any)();
    expect(deleteResource).toHaveBeenCalledWith(defaultProps.namespace, "foo", resource);
    // wait async calls
    await new Promise(r => r());
    expect(push).toHaveBeenCalledWith(app.apps.list(defaultProps.cluster, defaultProps.namespace));
  });

  it("updates the state with the CRD resources", () => {
    const wrapper = shallow(<OperatorInstance {...defaultProps} />);
    wrapper.setProps({ csv, resource });
    expect(wrapper.state("resources")).toMatchObject({
      deployRefs: [
        {
          apiVersion: "apps/v1",
          filter: {
            metadata: {
              ownerReferences: [
                {
                  kind: "Foo",
                  name: "foo-instance",
                },
              ],
            },
          },
        },
      ],
    });
  });

  it("updates the state with all the resources if the CRD doesn't define any", () => {
    const wrapper = shallow(<OperatorInstance {...defaultProps} />);
    const csvWithoutResource = {
      ...csv,
      spec: {
        ...csv.spec,
        customresourcedefinitions: {
          owned: [
            {
              name: "foo.kubeapps.com",
              version: "v1alpha1",
              kind: "Foo",
            },
          ],
        },
      },
    } as any;
    wrapper.setProps({ csv: csvWithoutResource, resource });
    const resources = wrapper.state("resources") as IPartialAppViewState;
    const resourcesKeys = Object.keys(resources).filter(k => k !== "otherResources");
    resourcesKeys.forEach(k => expect(resources[k].length).toBe(1));
  });

  it("skips AppNotes and AppValues if the resource doesn't have spec or status", () => {
    const wrapper = shallow(<OperatorInstance {...defaultProps} />);
    wrapper.setProps({ csv, resource: { ...resource, spec: undefined, status: undefined } });
    expect(wrapper.find(AppNotes)).not.toExist();
    expect(wrapper.find(AppValues)).not.toExist();
  });
});
