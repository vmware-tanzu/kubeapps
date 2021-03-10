import { act } from "react-dom/test-utils";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { IChartState, IChartVersion } from "shared/types";
import BasicDeploymentForm from "./BasicDeploymentForm";
import DeploymenetFormBody, { IDeploymentFormBodyProps } from "./DeploymentFormBody";
import DifferentialSelector from "./DifferentialSelector";

const defaultProps: IDeploymentFormBodyProps = {
  deploymentEvent: "install",
  chartID: "foo",
  chartVersion: "1.0.0",
  chartsIsFetching: false,
  selected: {} as IChartState["selected"],
  appValues: "foo: bar\n",
  setValues: jest.fn(),
  setValuesModified: jest.fn(),
};

jest.useFakeTimers();

const versions = [
  { id: "foo", attributes: { version: "1.2.3" }, relationships: { chart: { data: { repo: {} } } } },
] as IChartVersion[];

// Note that most of the tests that cover DeploymentFormBody component are in
// in the DeploymentForm and UpgradeForm parent components

// Context at https://github.com/kubeapps/kubeapps/issues/1293
it("should modify the original values of the differential component if parsed as YAML object", () => {
  const oldValues = `a: b


c: d
`;
  const schema = { properties: { a: { type: "string", form: true } } };
  const selected = {
    values: oldValues,
    schema,
    versions: [versions[0], { ...versions[0], attributes: { version: "1.2.4" } } as IChartVersion],
    version: versions[0],
  };

  const wrapper = mountWrapper(
    defaultStore,
    <DeploymenetFormBody {...defaultProps} selected={selected} />,
  );
  expect(wrapper.find(DifferentialSelector).prop("defaultValues")).toBe(oldValues);

  // Trigger a change in the basic form and a YAML parse
  const input = wrapper.find(BasicDeploymentForm).find("input");
  act(() => {
    input.simulate("change", { currentTarget: "e" });
    jest.advanceTimersByTime(500);
  });
  wrapper.update();

  // The original double empty line gets deleted
  const expectedValues = `a: b

c: d
`;
  expect(wrapper.find(DifferentialSelector).prop("defaultValues")).toBe(expectedValues);
});
