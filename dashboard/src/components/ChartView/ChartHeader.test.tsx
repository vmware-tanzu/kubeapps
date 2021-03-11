import { mount } from "enzyme";

import ChartHeader from "./ChartHeader";

const testProps: any = {
  chartAttrs: {
    description: "A Test Chart",
    name: "test",
    repo: {
      name: "testrepo",
    },
    namespace: "kubeapps",
    cluster: "default",
    icon: "test.jpg",
  },
  versions: [
    {
      attributes: {
        app_version: "1.2.3",
      },
    },
  ],
  onSelect: jest.fn(),
};

it("renders a header for the chart", () => {
  const wrapper = mount(<ChartHeader {...testProps} />);
  expect(wrapper.text()).toContain("testrepo/test");
});

it("displays the appVersion", () => {
  const wrapper = mount(<ChartHeader {...testProps} />);
  expect(wrapper.text()).toContain("1.2.3");
});

it("uses the icon", () => {
  const wrapper = mount(<ChartHeader {...testProps} />);
  const icon = wrapper.find("img").filterWhere(i => i.prop("alt") === "icon");
  expect(icon.exists()).toBe(true);
  expect(icon.props()).toMatchObject({ src: "api/assetsvc/test.jpg" });
});

it("uses the first version as default in the select input", () => {
  const versions = [
    {
      attributes: {
        version: "1.2.3",
      },
    },
    {
      attributes: {
        version: "1.2.4",
      },
    },
  ];
  const wrapper = mount(<ChartHeader {...testProps} versions={versions} />);
  expect(wrapper.find("select").prop("value")).toBe("1.2.3");
});

it("uses the current version as default in the select input", () => {
  const versions = [
    {
      attributes: {
        version: "1.2.3",
      },
    },
    {
      attributes: {
        version: "1.2.4",
      },
    },
  ];
  const wrapper = mount(<ChartHeader {...testProps} versions={versions} currentVersion="1.2.4" />);
  expect(wrapper.find("select").prop("value")).toBe("1.2.4");
});
