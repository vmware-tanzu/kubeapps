// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import { Maintainer } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import AvailablePackageMaintainers from "./AvailablePackageMaintainers";

const tests: Array<{
  expectedLinks: Array<string | null>;
  githubIDAsNames?: boolean;
  maintainers: Maintainer[];
  name: string;
}> = [
  {
    expectedLinks: [null],
    maintainers: [{ name: "Test Author", email: "" }],
    name: "with no email",
  },
  {
    expectedLinks: [null, "mailto:test@example.com"],
    maintainers: [
      { name: "Test Author", email: "" },
      { name: "Test Author 2", email: "test@example.com" },
    ],
    name: "with email",
  },
  {
    expectedLinks: ["https://github.com/test1", "https://github.com/test2"],
    githubIDAsNames: true,
    maintainers: [
      { name: "test1", email: "" },
      { name: "test2", email: "test@example.com" },
    ],
    name: "with github ids",
  },
];

for (const t of tests) {
  it(`it renders the maintainers list ${t.name}`, () => {
    const wrapper = shallow(
      <AvailablePackageMaintainers
        maintainers={t.maintainers}
        githubIDAsNames={t.githubIDAsNames}
      />,
    );
    const list = wrapper.find("li");
    expect(list).toHaveLength(t.maintainers.length);
    list.forEach((li, i) => {
      if (t.expectedLinks[i]) {
        expect(li.find("a").props()).toMatchObject({ href: t.expectedLinks[i] });
      } else {
        expect(li.props().children).toBe(t.maintainers[i].name);
      }
      expect(li.text()).toBe(t.maintainers[i].name);
    });
  });
}
