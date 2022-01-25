// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import React from "react";
import Column from ".";

const randomColumn = () => Math.floor(Math.random() * Math.floor(11)) + 1;

const randomColumns = numberOfColumns =>
  [...Array(numberOfColumns).keys()].map(() => randomColumn());

describe(Column, () => {
  it("renders the column with auto when any property is passed", () => {
    const wrapper = shallow(<Column>Test</Column>);

    expect(wrapper.prop("className")).toBe("clr-col");
  });

  it("adds the correct role when it's a list item", () => {
    const wrapper = shallow(<Column listItem>Test</Column>);

    expect(wrapper.prop("role")).toBe("listitem");
  });

  describe("Fixed properties", () => {
    it("adds the clr-col-X class when span is a number", () => {
      const span = 6;
      const wrapper = shallow(<Column span={span}>Test</Column>);

      expect(wrapper.prop("className")).toBe(`clr-col-${span}`);
    });

    it("adds the clr-offset-X claoffset when offset is a number", () => {
      const offset = 6;
      const wrapper = shallow(<Column offset={offset}>Test</Column>);

      expect(wrapper.prop("className")).toBe(`clr-offset-${offset}`);
    });

    it("adds clr-col and clr-offset", () => {
      const span = 6;
      const offset = 6;
      const wrapper = shallow(
        <Column span={span} offset={offset}>
          Test
        </Column>,
      );

      expect(wrapper).toHaveClassName(`clr-col-${span}`);
      expect(wrapper).toHaveClassName(`clr-offset-${offset}`);
    });
  });

  describe("Responsive Properties", () => {
    it("adds the clr-col responsive columns when span is an array of 1 element", () => {
      const span = randomColumns(1);
      const wrapper = shallow(<Column span={span}>Test</Column>);

      expect(wrapper.prop("className")).toBe(`clr-col-sm-${span[0]}`);
    });

    it("adds the clr-col responsive columns when span is an array of 2 element", () => {
      const span = randomColumns(2);
      const wrapper = shallow(<Column span={span}>Test</Column>);

      expect(wrapper).toHaveClassName(`clr-col-sm-${span[0]}`);
      expect(wrapper).toHaveClassName(`clr-col-md-${span[1]}`);
    });

    it("adds the clr-col responsive columns when span is an array of 3 element", () => {
      const span = randomColumns(3);
      const wrapper = shallow(<Column span={span}>Test</Column>);

      expect(wrapper).toHaveClassName(`clr-col-sm-${span[0]}`);
      expect(wrapper).toHaveClassName(`clr-col-md-${span[1]}`);
      expect(wrapper).toHaveClassName(`clr-col-lg-${span[2]}`);
    });

    it("adds the clr-col responsive columns when span is an array of 4 element", () => {
      const span = randomColumns(4);
      const wrapper = shallow(<Column span={span}>Test</Column>);

      expect(wrapper).toHaveClassName(`clr-col-sm-${span[0]}`);
      expect(wrapper).toHaveClassName(`clr-col-md-${span[1]}`);
      expect(wrapper).toHaveClassName(`clr-col-lg-${span[2]}`);
      expect(wrapper).toHaveClassName(`clr-col-xl-${span[3]}`);
    });

    it("adds the clr-col responsive columns and ignore undefined responsive breakpoints", () => {
      const span = randomColumns(5);
      const wrapper = shallow(<Column span={span}>Test</Column>);

      expect(wrapper).toHaveClassName(`clr-col-sm-${span[0]}`);
      expect(wrapper).toHaveClassName(`clr-col-md-${span[1]}`);
      expect(wrapper).toHaveClassName(`clr-col-lg-${span[2]}`);
      expect(wrapper).toHaveClassName(`clr-col-xl-${span[3]}`);
    });

    it("adds the clr-offset responsive offset when offset is an array of 1 element", () => {
      const offset = randomColumns(1);
      const wrapper = shallow(<Column offset={offset}>Test</Column>);

      expect(wrapper.prop("className")).toBe(`clr-offset-sm-${offset[0]}`);
    });

    it("adds the clr-offset responsive offset when offset is an array of 2 element", () => {
      const offset = randomColumns(2);
      const wrapper = shallow(<Column offset={offset}>Test</Column>);

      expect(wrapper).toHaveClassName(`clr-offset-sm-${offset[0]}`);
      expect(wrapper).toHaveClassName(`clr-offset-md-${offset[1]}`);
    });

    it("adds the clr-offset responsive offset when offset is an array of 3 element", () => {
      const offset = randomColumns(3);
      const wrapper = shallow(<Column offset={offset}>Test</Column>);

      expect(wrapper).toHaveClassName(`clr-offset-sm-${offset[0]}`);
      expect(wrapper).toHaveClassName(`clr-offset-md-${offset[1]}`);
      expect(wrapper).toHaveClassName(`clr-offset-lg-${offset[2]}`);
    });

    it("adds the clr-offset responsive offset when offset is an array of 4 element", () => {
      const offset = randomColumns(4);
      const wrapper = shallow(<Column offset={offset}>Test</Column>);

      expect(wrapper).toHaveClassName(`clr-offset-sm-${offset[0]}`);
      expect(wrapper).toHaveClassName(`clr-offset-md-${offset[1]}`);
      expect(wrapper).toHaveClassName(`clr-offset-lg-${offset[2]}`);
      expect(wrapper).toHaveClassName(`clr-offset-xl-${offset[3]}`);
    });

    it("adds the clr-offset responsive offset and ignore undefined responsive breakpoints", () => {
      const offset = randomColumns(5);
      const wrapper = shallow(<Column offset={offset}>Test</Column>);

      expect(wrapper).toHaveClassName(`clr-offset-sm-${offset[0]}`);
      expect(wrapper).toHaveClassName(`clr-offset-md-${offset[1]}`);
      expect(wrapper).toHaveClassName(`clr-offset-lg-${offset[2]}`);
      expect(wrapper).toHaveClassName(`clr-offset-xl-${offset[3]}`);
    });

    it("combine both responsive properties", () => {
      const offset = randomColumns(5);
      const span = randomColumns(5);
      const wrapper = shallow(
        <Column span={span} offset={offset}>
          Test
        </Column>,
      );

      expect(wrapper).toHaveClassName(`clr-col-sm-${span[0]}`);
      expect(wrapper).toHaveClassName(`clr-col-md-${span[1]}`);
      expect(wrapper).toHaveClassName(`clr-col-lg-${span[2]}`);
      expect(wrapper).toHaveClassName(`clr-col-xl-${span[3]}`);
      expect(wrapper).toHaveClassName(`clr-offset-sm-${offset[0]}`);
      expect(wrapper).toHaveClassName(`clr-offset-md-${offset[1]}`);
      expect(wrapper).toHaveClassName(`clr-offset-lg-${offset[2]}`);
      expect(wrapper).toHaveClassName(`clr-offset-xl-${offset[3]}`);
    });
  });
});
