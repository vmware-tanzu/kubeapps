import { axiosWithAuth } from "./AxiosInstance";
import Secret from "./Secret";

it("creates a secret", async () => {
  axiosWithAuth.post = jest.fn(() => {
    return { data: "ok" };
  });
  const secrets = {
    foo: "bar",
  };
  const owner = {
    foo: "bar",
  } as any;
  const name = "secret";
  const namespace = "default";
  expect(await Secret.create(name, secrets, owner, namespace)).toEqual("ok");
  expect(axiosWithAuth.post).toHaveBeenCalledWith("api/kube/api/v1/namespaces/default/secrets", {
    apiVersion: "v1",
    data: secrets,
    kind: "Secret",
    metadata: { name: "secret", ownerReferences: [owner] },
    type: "Opaque",
  });
});

it("deletes a secret", async () => {
  axiosWithAuth.delete = jest.fn();
  await Secret.delete("foo", "bar");
  expect(axiosWithAuth.delete).toHaveBeenCalledWith("api/kube/api/v1/namespaces/bar/secrets/foo");
});

it("gets a secret", async () => {
  axiosWithAuth.get = jest.fn(() => {
    return { data: "ok" };
  });
  await Secret.get("foo", "bar");
  expect(axiosWithAuth.get).toHaveBeenCalledWith("api/kube/api/v1/namespaces/bar/secrets/foo");
});

it("lists secrets", async () => {
  axiosWithAuth.get = jest.fn(() => {
    return { data: "ok" };
  });
  await Secret.list("foo");
  expect(axiosWithAuth.get).toHaveBeenCalledWith("api/kube/api/v1/namespaces/foo/secrets");
});

it("creates a pull secret", async () => {
  axiosWithAuth.post = jest.fn(() => {
    return { data: "ok" };
  });
  const name = "repo-1";
  const user = "foo";
  const password = "pass";
  const email = "foo@bar.com";
  const server = "docker.io";
  const namespace = "default";
  expect(await Secret.createPullSecret(name, user, password, email, server, namespace)).toBe("ok");
  expect(axiosWithAuth.post).toHaveBeenCalledWith("api/kube/api/v1/namespaces/default/secrets", {
    apiVersion: "v1",
    stringData: {
      ".dockerconfigjson":
        '{"auths":{"docker.io":{"username":"foo","password":"pass","email":"foo@bar.com","auth":"Zm9vOnBhc3M="}}}',
    },
    kind: "Secret",
    metadata: { name: "repo-1" },
    type: "kubernetes.io/dockerconfigjson",
  });
});
