import actions from "actions";
import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper.v2";
import SearchFilter from "components/SearchFilter/SearchFilter.v2";
import * as React from "react";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IPackageManifest } from "../../shared/types";
import { CardGrid } from "../Card";
import InfoCard from "../InfoCard/InfoCard.v2";
import { AUTO_PILOT, BASIC_INSTALL } from "../OperatorView/OperatorCapabilityLevel";
import OLMNotFound from "./OLMNotFound.v2";
import OperatorList, { IOperatorListProps } from "./OperatorList.v2";
import OperatorNotSupported from "./OperatorsNotSupported.v2";

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.operators };
beforeEach(() => {
  actions.operators = {
    ...actions.operators,
    checkOLMInstalled: jest.fn(),
    getOperators: jest.fn(),
    getCSVs: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.operators = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

const defaultProps: IOperatorListProps = {
  cluster: initialState.config.kubeappsCluster,
  namespace: "default",
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

it("renders a LoadingWrapper if fetching", () => {
  const wrapper = mountWrapper(
    getStore({ operators: { isFetcing: true } }),
    <OperatorList {...defaultProps} />,
  );
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("call the OLM check and render the NotFound message if not found", () => {
  const checkOLMInstalled = jest.fn();
  actions.operators.checkOLMInstalled = checkOLMInstalled;
  const wrapper = mountWrapper(defaultStore, <OperatorList {...defaultProps} />);
  expect(checkOLMInstalled).toHaveBeenCalled();
  expect(wrapper.find(OLMNotFound)).toExist();
});

it("displays an alert if rendered for an additional cluster", () => {
  const props = { ...defaultProps, cluster: "other-cluster" };
  const wrapper = mountWrapper(defaultStore, <OperatorList {...props} />);
  expect(wrapper.find(OperatorNotSupported)).toExist();
});

it("renders an error", () => {
  const wrapper = mountWrapper(
    getStore({ operators: { errors: { operator: { fetch: new Error("Forbidden!") } } } }),
    <OperatorList {...defaultProps} />,
  );
  const error = wrapper.find(Alert).filterWhere(a => a.prop("theme") === "danger");
  expect(error).toExist();
  expect(error).toIncludeText("Forbidden!");
});

it("request operators if the OLM is installed", () => {
  const getOperators = jest.fn();
  actions.operators.getOperators = getOperators;
  const wrapper = mountWrapper(
    getStore({ operators: { isOLMInstalled: true } }),
    <OperatorList {...defaultProps} />,
  );
  wrapper.setProps({ namespace: "other" });
  expect(getOperators).toHaveBeenCalled();
});

it("render the operator list", () => {
  const wrapper = mountWrapper(
    getStore({ operators: { isOLMInstalled: true, operators: [sampleOperator] } }),
    <OperatorList {...defaultProps} />,
  );
  expect(wrapper.find(OLMNotFound)).not.toExist();
  expect(wrapper.find(InfoCard)).toExist();
});

// TODO(andresmgot): Enable when the list of installed vs available is ready
xit("render the operator list with installed operators", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: { isOLMInstalled: true, operators: [sampleOperator], csvs: [sampleCSV] },
    }),
    <OperatorList {...defaultProps} />,
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

xit("render the operator list without installed operators", () => {
  const wrapper = mountWrapper(
    getStore({ operators: { isOLMInstalled: true, operators: [sampleOperator] } }),
    <OperatorList {...defaultProps} />,
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
    const wrapper = mountWrapper(
      getStore({
        operators: { isOLMInstalled: true, operators: [sampleOperator, sampleOperator2] },
      }),
      <OperatorList {...defaultProps} />,
    );
    expect(wrapper.find(InfoCard).length).toBe(2);
    act(() => {
      (wrapper.find(SearchFilter).prop("onChange") as any)("foo");
    });
    wrapper.update();
    const operator = wrapper.find(InfoCard);
    expect(operator.length).toBe(1);
    expect(operator.prop("title")).toBe(sampleOperator.metadata.name);
  });

  it("setting the filter in the props", () => {
    const wrapper = mountWrapper(
      getStore({
        operators: { isOLMInstalled: true, operators: [sampleOperator, sampleOperator2] },
      }),
      <OperatorList {...defaultProps} filter="foo" />,
    );
    const operator = wrapper.find(InfoCard);
    expect(operator.length).toBe(1);
    expect(operator.prop("title")).toBe(sampleOperator.metadata.name);
  });

  it("show a message if the filter doesn't match any operator", () => {
    const wrapper = mountWrapper(
      getStore({
        operators: { isOLMInstalled: true, operators: [sampleOperator, sampleOperator2] },
      }),
      <OperatorList {...defaultProps} filter="nope" />,
    );
    expect(wrapper.find(InfoCard)).not.toExist();
    expect(wrapper).toIncludeText("No operator matches the current filter");
  });

  it("filters by category", () => {
    const wrapper = mountWrapper(
      getStore({
        operators: { isOLMInstalled: true, operators: [sampleOperator, sampleOperator2] },
      }),
      <OperatorList {...defaultProps} />,
    );
    expect(wrapper.find(InfoCard).length).toBe(2);

    // Filter category "security"
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "security");
    input.simulate("change", { target: { value: "security" } });
    const operator = wrapper.find(InfoCard);
    expect(operator.length).toBe(1);
    expect(operator.prop("title")).toBe(sampleOperator.metadata.name);
  });

  it("filters by capability", () => {
    const wrapper = mountWrapper(
      getStore({
        operators: { isOLMInstalled: true, operators: [sampleOperator, sampleOperator2] },
      }),
      <OperatorList {...defaultProps} />,
    );
    expect(wrapper.find(InfoCard).length).toBe(2);

    // Filter by capability "Basic Install"
    const input = wrapper.find("input").findWhere(i => i.prop("value") === BASIC_INSTALL);
    input.simulate("change", { target: { value: BASIC_INSTALL } });
    const operator = wrapper.find(InfoCard);
    expect(operator.length).toBe(1);
    expect(operator.prop("title")).toBe(sampleOperator2.metadata.name);
  });
});
