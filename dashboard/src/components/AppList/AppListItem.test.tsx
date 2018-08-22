import { shallow } from "enzyme";
import * as React from "react";
import { Link } from "react-router-dom";

import { IAppOverview } from "../../shared/types";
import Card from "../Card";
import AppListItem from "./AppListItem";

it("renders a app item", () => {
  const wrapper = shallow(
    <AppListItem
      app={
        {
          namespace: "default",
          releaseName: "foo",
          status: "DEPLOYED",
          version: "1.0.0",
        } as IAppOverview
      }
    />,
  );
  expect(wrapper.find(Card).key()).toBe("foo");
  expect(
    wrapper
      .find(Card)
      .children()
      .find(Link)
      .props().title,
  ).toBe("foo");
  expect(
    wrapper
      .find(Card)
      .children()
      .find(Link)
      .props().to,
  ).toBe("/apps/ns/default/foo");
  expect(wrapper.find(".ChartListItem__content__info_version").text()).toBe("1.0.0");
  expect(wrapper.find(".DEPLOYED").exists()).toBe(true);
  expect(wrapper.find(".ChartListItem__content__info_repo").text()).toBe("default");
  expect(wrapper.find(".ChartListItem__content__info_other").text()).toBe("deployed");
});
