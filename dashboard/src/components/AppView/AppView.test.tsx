import { shallow } from "enzyme";
import * as React from "react";
import { hapi } from "../../shared/hapi/release";
import { IResource } from "../../shared/types";
import AppViewComponent, { IAppViewProps } from "./AppView";
import manifest from "./testdata/test-k8s-manifest";

describe("AppViewComponent", () => {
  let validProps: IAppViewProps;
  beforeEach(() => {
    const appRelease = hapi.release.Release.create({
      info: hapi.release.Info.create(),
      manifest,
      namespace: "weee",
    });

    validProps = {
      app: appRelease,
      deleteApp: jest.fn(),
      deleteError: undefined,
      error: undefined,
      getApp: jest.fn(),
      namespace: "my-happy-place",
      releaseName: "mr-sunshine",
    };
  });

  it("renders a loading message if info is not present", () => {
    validProps.app.info = null;
    const wrapper = shallow(<AppViewComponent {...validProps} />);
    expect(wrapper.text()).toBe("Loading");
  });

  describe("State initialization", () => {
    /*
      The imported manifest contains one deployment, one service, one config map and some bogus manifests.
      We only set websockets for deployment and services
    */
    it("sets a list of web sockets for its deployments and services", () => {
      const wrapper = shallow(<AppViewComponent {...validProps} />);
      wrapper.setProps(validProps);
      const sockets: WebSocket[] = wrapper.state("sockets");
      expect(sockets.length).toEqual(2);
      expect(sockets[0].url).toBe(
        "ws://localhost/api/kube/apis/apps/v1beta1/namespaces/weee/deployments?watch=true&fieldSelector=metadata.name%3Ddeployment-one",
      );
      expect(sockets[1].url).toBe(
        "ws://localhost/api/kube/api/v1/namespaces/weee/services?watch=true&fieldSelector=metadata.name%3Dsvc-one",
      );
    });

    it("stores other k8s resources directly in the state", () => {
      const wrapper = shallow(<AppViewComponent {...validProps} />);
      wrapper.setProps(validProps);
      const otherResources: Map<string, IResource> = wrapper.state("otherResources");
      const configMap = otherResources["ConfigMap/cm-one"];
      // Only one element in the otherResources array
      // This means that no other spurious bogus (i.e missing kind: ) definitions were added.
      expect(Object.keys(otherResources).length).toEqual(1);
      // The config map is stored
      expect(configMap).toBeDefined();
      expect(configMap.metadata.name).toEqual("cm-one");
    });
  });
});
