import { shallow } from "enzyme";
import * as React from "react";

import { IChartAttributes } from "../../shared/types";
import ChartMaintainers from "./ChartMaintainers";

const tests: Array<{
  expectedLinks: Array<string | null>;
  githubIDAsNames?: boolean;
  maintainers: IChartAttributes["maintainers"];
  name: string;
}> = [
  {
    expectedLinks: [null],
    maintainers: [{ name: "Test Author" }],
    name: "with no email",
  },
  {
    expectedLinks: [null, "mailto:test@example.com"],
    maintainers: [{ name: "Test Author" }, { name: "Test Author 2", email: "test@example.com" }],
    name: "with email",
  },
  {
    expectedLinks: ["https://github.com/test1", "https://github.com/test2"],
    githubIDAsNames: true,
    maintainers: [{ name: "test1" }, { name: "test2", email: "test@example.com" }],
    name: "with github ids",
  },
];

for (const t of tests) {
  it(`it renders the maintainers list ${t.name}`, () => {
    const wrapper = shallow(
      <ChartMaintainers maintainers={t.maintainers} githubIDAsNames={t.githubIDAsNames} />,
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
