import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import { cloneDeep } from "lodash";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import OperatorSummary from "./OperatorSummary";

const defaultOperator = {
  metadata: {
    name: "foo",
    namespace: "kubeapps",
  },
  status: {
    provider: {
      name: "Kubeapps",
    },
    defaultChannel: "beta",
    channels: [
      {
        name: "beta",
        currentCSV: "foo.1.0.0",
        currentCSVDesc: {
          displayName: "Foo",
          version: "1.0.0",
          description: "this is a testing operator",
          annotations: {
            capabilities: "Basic Install",
            repository: "github.com/kubeapps/kubeapps",
            containerImage: "kubeapps/kubeapps",
            createdAt: "one day",
          },
          installModes: [],
        },
      },
    ],
  },
} as any;

it("shows a loading wrapper", () => {
  const wrapper = mountWrapper(getStore({ operators: { isFetching: true } }), <OperatorSummary />);
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("shows an alert if the operator doesn't have a channel", () => {
  const operatorWithoutChannel = cloneDeep(defaultOperator);
  operatorWithoutChannel.status.channels = [];
  const wrapper = mountWrapper(
    getStore({ operators: { operator: operatorWithoutChannel } }),
    <OperatorSummary />,
  );
  expect(wrapper.find(Alert)).toExist();
});

it("doesn't fail with missing info", () => {
  const operatorWithoutAnnotations = cloneDeep(defaultOperator);
  delete operatorWithoutAnnotations.status.channels[0].currentCSVDesc.annotations;
  const wrapper = mountWrapper(
    getStore({ operators: { operator: operatorWithoutAnnotations } }),
    <OperatorSummary />,
  );
  expect(wrapper.find(".left-menu")).toExist();
});

it("shows all the operator info", () => {
  const wrapper = mountWrapper(
    getStore({ operators: { operator: defaultOperator } }),
    <OperatorSummary />,
  );
  expect(wrapper.find(".left-menu-subsection")).toHaveLength(5);
});
