import { shallow } from "enzyme";
import * as React from "react";
import { ISecret } from "shared/types";
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
