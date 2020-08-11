import { shallow } from "enzyme";
import * as React from "react";
import itBehavesLike from "../../shared/specs";
import { ForbiddenError, IPackageManifest } from "../../shared/types";
import { CardGrid } from "../Card";
import { ErrorSelector } from "../ErrorAlert";
import InfoCard from "../InfoCard";
import LoadingWrapper from "../LoadingWrapper";
import { AUTO_PILOT, BASIC_INSTALL } from "../OperatorView/OperatorCapabilityLevel";
import OLMNotFound from "./OLMNotFound";
import OperatorList, { IOperatorListProps } from "./OperatorList";
import OperatorNotSupported from "./OperatorsNotSupported";

const defaultProps: IOperatorListProps = {
  isFetching: false,
  checkOLMInstalled: jest.fn(),
  isOLMInstalled: false,
  operators: [],
  cluster: "default",
  namespace: "default",
  getOperators: jest.fn(),
  getCSVs: jest.fn(),
  csvs: [],
  filter: "",
  pushSearchFilter: jest.fn(),
};

const sampleOperator = {
  metadata: {
    name: "foo",
  },
  status: {
    provider: {
      name: "kubeapps",
    },
    defaultChannel: "alpha",
    channels: [
      {
        name: "alpha",
        currentCSV: "kubeapps-operator",
        currentCSVDesc: {
          version: "1.0.0",
          annotations: {
            categories: "security",
            capabilities: AUTO_PILOT,
          },
        },
      },
    ],
  },
} as IPackageManifest;

const sampleCSV = {
  metadata: { name: "kubeapps-operator" },
  spec: {
    icon: [{}],
    provider: {
      name: "kubeapps",
    },
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

itBehavesLike("aLoadingComponent", {
  component: OperatorList,
  props: { ...defaultProps, isFetching: true },
});

it("call the OLM check and render the NotFound message if not found", () => {
  const checkOLMInstalled = jest.fn();
  const wrapper = shallow(<OperatorList {...defaultProps} checkOLMInstalled={checkOLMInstalled} />);
  expect(checkOLMInstalled).toHaveBeenCalled();
  expect(wrapper.find(OLMNotFound)).toExist();
});

it("displays an alert if rendered for an additional cluster", () => {
  const props = { ...defaultProps, cluster: "other-cluster" };
  const wrapper = shallow(<OperatorList {...props} />);
  expect(wrapper.find(OperatorNotSupported)).toExist();
});

it("call the OLM check and render the a Forbidden message if it could do the check", () => {
  const checkOLMInstalled = jest.fn();
  const wrapper = shallow(<OperatorList {...defaultProps} checkOLMInstalled={checkOLMInstalled} />);
  wrapper.setProps({ error: new ForbiddenError("nope") });
  expect(checkOLMInstalled).toHaveBeenCalled();
  expect(wrapper.find(ErrorSelector)).toExist();
  expect(wrapper.find(OLMNotFound)).not.toExist();
});

it("re-request operators if the namespace changes", () => {
  const getOperators = jest.fn();
  const getCSVs = jest.fn();
  const wrapper = shallow(
    <OperatorList {...defaultProps} getOperators={getOperators} getCSVs={getCSVs} />,
  );
  wrapper.setProps({ namespace: "other" });
  expect(getOperators).toHaveBeenCalledTimes(2);
  expect(getCSVs).toHaveBeenCalledTimes(2);
});

it("renders an error if exists", () => {
  const wrapper = shallow(
    <OperatorList {...defaultProps} isOLMInstalled={true} error={new Error("Boom!")} />,
  );
  expect(wrapper.find(ErrorSelector)).toExist();
  expect(
    wrapper
      .find(ErrorSelector)
      .dive()
      .dive()
      .text(),
  ).toMatch("Boom!");
});

it("skips the error if the OLM is not installed", () => {
  const wrapper = shallow(
    <OperatorList
      {...defaultProps}
      isOLMInstalled={false}
      error={new Error("There are no operators!")}
    />,
  );
  expect(wrapper.find(ErrorSelector)).not.toExist();
  expect(wrapper.find(OLMNotFound)).toExist();
});

it("render the operator list with installed operators", () => {
  const wrapper = shallow(
    <OperatorList
      {...defaultProps}
      isOLMInstalled={true}
      operators={[sampleOperator]}
      csvs={[sampleCSV]}
    />,
  );
  expect(wrapper.find(OLMNotFound)).not.toExist();
  expect(wrapper.find(InfoCard)).toExist();
  // The section "Available operators" should be empty since all the ops are installed
  expect(wrapper.find("h3").filterWhere(c => c.text() === "Installed")).toExist();
  expect(
    wrapper
      .find(CardGrid)
      .last()
      .children(),
  ).not.toExist();
  expect(wrapper).toMatchSnapshot();
});

it("render the operator list without installed operators", () => {
  const wrapper = shallow(
    <OperatorList {...defaultProps} isOLMInstalled={true} operators={[sampleOperator]} csvs={[]} />,
  );
  expect(wrapper.find(OLMNotFound)).not.toExist();
  expect(wrapper.find(InfoCard)).toExist();
  // The section "Available operators" should not be empty since the operator is not installed
  expect(wrapper.find("h3").filterWhere(c => c.text() === "Installed")).not.toExist();
  expect(
    wrapper
      .find(CardGrid)
      .last()
      .children(),
  ).toExist();
  expect(wrapper).toMatchSnapshot();
});

describe("filter operators", () => {
  const sampleOperator2 = {
    ...sampleOperator,
    metadata: {
      name: "bar",
    },
    status: {
      ...sampleOperator.status,
      channels: [
        {
          ...sampleOperator.status.channels[0],
          currentCSVDesc: {
            version: "1.0.0",
            annotations: {
              categories: "database, other",
              capabilities: BASIC_INSTALL,
            },
          },
        },
      ],
    },
  } as any;

  it("setting the filter in the state", () => {
    const wrapper = shallow(
      <OperatorList
        {...defaultProps}
        isOLMInstalled={true}
        operators={[sampleOperator, sampleOperator2]}
        csvs={[]}
      />,
    );
    expect(wrapper.find(InfoCard).length).toBe(2);
    wrapper.setState({ filter: "foo" });
    expect(wrapper.find(InfoCard).length).toBe(1);
  });

  it("setting the filter in the props", () => {
    const wrapper = shallow(
      <OperatorList
        {...defaultProps}
        isOLMInstalled={true}
        operators={[sampleOperator, sampleOperator2]}
        csvs={[]}
      />,
    );
    expect(wrapper.find(InfoCard).length).toBe(2);
    wrapper.setProps({ filter: "foo" });
    expect(wrapper.find(InfoCard).length).toBe(1);
  });

  it("show a message if the filter doesn't match any operator", () => {
    const wrapper = shallow(
      <OperatorList
        {...defaultProps}
        isOLMInstalled={true}
        operators={[sampleOperator, sampleOperator2]}
        csvs={[]}
      />,
    );
    expect(wrapper.find(InfoCard).length).toBe(2);
    wrapper.setProps({ filter: "nope" });
    expect(wrapper.find(InfoCard)).not.toExist();
    expect(
      wrapper
        .find(LoadingWrapper)
        .dive()
        .text(),
    ).toMatch("No Operator found");
    expect(wrapper.find(".horizontal-column")).toExist();
  });

  it("filters by category", () => {
    const wrapper = shallow(<OperatorList {...defaultProps} isOLMInstalled={true} csvs={[]} />);
    wrapper.setProps({ operators: [sampleOperator, sampleOperator2] });
    const column = wrapper.find(".horizontal-column").text();
    expect(column).toContain("security");
    expect(column).toContain("database");
    expect(column).toContain("other");
    expect(wrapper.find(InfoCard).length).toBe(2);

    // Filter category "security"
    wrapper.setState({
      filterCategories: {
        security: true,
      },
    });
    expect(wrapper.find(InfoCard).length).toBe(1);
  });

  it("filters by capability", () => {
    const wrapper = shallow(<OperatorList {...defaultProps} isOLMInstalled={true} csvs={[]} />);
    wrapper.setProps({ operators: [sampleOperator, sampleOperator2] });
    expect(wrapper.find(InfoCard).length).toBe(2);

    // Filter by capability "Basic Install"
    wrapper.setState({
      filterCapabilities: {
        [BASIC_INSTALL]: true,
      },
    });
    expect(wrapper.find(InfoCard).length).toBe(1);
  });
});
