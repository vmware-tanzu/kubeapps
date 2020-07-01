import { shallow } from "enzyme";
import * as React from "react";
import { ISecret } from "shared/types";
import { wait } from "../../../shared/utils";
import AppRepoAddDockerCreds from "./AppRepoAddDockerCreds";

const secret1 = {
  metadata: {
    name: "foo",
  },
} as ISecret;
const secret2 = {
  metadata: {
    name: "bar",
  },
} as ISecret;
const defaultProps = {
  imagePullSecrets: [],
  togglePullSecret: jest.fn(),
  selectedImagePullSecrets: {},
  namespace: "default",
  createDockerRegistrySecret: jest.fn(),
  fetchImagePullSecrets: jest.fn(),
};

it("shows an info message if there are no secrets", () => {
  const wrapper = shallow(<AppRepoAddDockerCreds {...defaultProps} />);
  expect(wrapper.text()).toContain("No existing credentials found");
});

it("shows the list of available pull secrets", () => {
  const wrapper = shallow(
    <AppRepoAddDockerCreds {...defaultProps} imagePullSecrets={[secret1, secret2]} />,
  );
  expect(wrapper.text()).toContain(secret1.metadata.name);
  expect(wrapper.text()).toContain(secret2.metadata.name);
});

it("select secrets", () => {
  const wrapper = shallow(
    <AppRepoAddDockerCreds
      {...defaultProps}
      imagePullSecrets={[secret1, secret2]}
      selectedImagePullSecrets={{ [secret1.metadata.name]: true }}
    />,
  );
  const totalCheckbox = wrapper.find("input").filterWhere(i => i.prop("type") === "checkbox");
  expect(totalCheckbox.length).toBe(2);

  const selectedCheckbox = totalCheckbox.filterWhere(i => i.prop("checked") === true);
  expect(selectedCheckbox.length).toBe(1);
});

it("renders the form to create a registry secret", () => {
  const wrapper = shallow(<AppRepoAddDockerCreds {...defaultProps} />);

  expect(wrapper.text()).not.toContain("Secret Name");

  const link = wrapper.find("button").filterWhere(b => b.text().includes("Add new"));
  link.simulate("click");

  expect(wrapper.text()).toContain("Secret Name");
});

it("submits the new secret and re-request the list", async () => {
  const fetchImagePullSecrets = jest.fn();
  const createDockerRegistrySecret = jest.fn().mockReturnValue(true);
  const wrapper = shallow(
    <AppRepoAddDockerCreds
      {...defaultProps}
      fetchImagePullSecrets={fetchImagePullSecrets}
      createDockerRegistrySecret={createDockerRegistrySecret}
    />,
  );
  const secretName = "repo-1";
  const user = "foo";
  const password = "pass";
  const email = "foo@bar.com";
  const server = "docker.io";
  wrapper.setState({ secretName, user, password, email, server, showSecretSubForm: true });

  const button = wrapper.find("button").filterWhere(a => a.text().includes("Submit"));
  button.simulate("click");

  await wait(1);
  expect(fetchImagePullSecrets).toHaveBeenCalledWith(defaultProps.namespace);
  expect(createDockerRegistrySecret).toHaveBeenCalledWith(
    secretName,
    user,
    password,
    email,
    server,
    defaultProps.namespace,
  );
});
