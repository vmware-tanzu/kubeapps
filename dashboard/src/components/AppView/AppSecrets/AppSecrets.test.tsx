import ResourceRef from "shared/ResourceRef";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { ISecret } from "shared/types";
import SecretItemDatum from "../ResourceTable/ResourceItem/SecretItem/SecretItemDatum";
import AppSecrets from "./AppSecrets";

const defaultProps = {
  secretRefs: [],
};

const sampleResourceRef = {
  cluster: "cluster-name",
  apiVersion: "v1",
  kind: "Secret",
  name: "foo",
  namespace: "default",
  filter: "",
  plural: "secrets",
  namespaced: true,
  getResourceURL: jest.fn(() => "secret-foo"),
  watchResourceURL: jest.fn(),
  getResource: jest.fn(),
  watchResource: jest.fn(),
} as ResourceRef;

const secret = {
  metadata: {
    name: "foo",
  },
  type: "Opaque",
  data: {
    foo: "YmFy", // bar
    foo2: "YmFyMg==", // bar2
  } as { [s: string]: string },
} as ISecret;

it("shows a message if there are no secrets", () => {
  const wrapper = mountWrapper(defaultStore, <AppSecrets {...defaultProps} />);
  expect(wrapper.text()).toContain("The current application does not include secrets");
});

it("renders a secretItemDatum per secret", () => {
  const state = getStore({
    kube: {
      items: {
        "secret-foo": {
          isFetching: false,
          item: secret,
        },
      },
    },
  });
  const wrapper = mountWrapper(state, <AppSecrets secretRefs={[sampleResourceRef]} />);
  expect(wrapper.find(SecretItemDatum)).toHaveLength(2);
});
