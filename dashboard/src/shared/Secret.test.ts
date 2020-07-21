import { axiosWithAuth } from "./AxiosInstance";
import Secret from "./Secret";

it("gets a secret", async () => {
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: "ok" });
  await Secret.get("default", "foo", "bar");
  expect(axiosWithAuth.get).toHaveBeenCalledWith(
    "api/clusters/default/api/v1/namespaces/bar/secrets/foo",
  );
});

it("lists secrets", async () => {
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: "ok" });
  await Secret.list("default", "foo");
  expect(axiosWithAuth.get).toHaveBeenCalledWith(
    "api/clusters/default/api/v1/namespaces/foo/secrets",
  );
});

it("creates a pull secret", async () => {
  axiosWithAuth.post = jest.fn().mockReturnValue({ data: "ok" });
  const name = "repo-1";
  const user = "foo";
  const password = "pass";
  const email = "foo@bar.com";
  const server = "docker.io";
  const namespace = "default";
  expect(
    await Secret.createPullSecret("default", name, user, password, email, server, namespace),
  ).toBe("ok");
  expect(axiosWithAuth.post).toHaveBeenCalledWith(
    "api/clusters/default/api/v1/namespaces/default/secrets",
    {
      apiVersion: "v1",
      stringData: {
        ".dockerconfigjson":
          '{"auths":{"docker.io":{"username":"foo","password":"pass","email":"foo@bar.com","auth":"Zm9vOnBhc3M="}}}',
      },
      kind: "Secret",
      metadata: { name: "repo-1" },
      type: "kubernetes.io/dockerconfigjson",
    },
  );
});
