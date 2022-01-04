import { axiosWithAuth } from "./AxiosInstance";
import Secret from "./Secret";
import {
  CreateSecretRequest,
  CreateSecretResponse,
  SecretType,
} from "gen/kubeappsapis/plugins/resources/v1alpha1/resources";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";

it("gets a secret", async () => {
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: "ok" });
  await Secret.get("default", "bar", "foo");
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

describe("createSecret", () => {
  // Create a real client, but we'll stub out the function we're interested in.
  const client = new KubeappsGrpcClient().getResourcesServiceClientImpl();
  let mockClientCreateSecret: jest.MockedFunction<typeof client.CreateSecret>;
  beforeEach(() => {
    mockClientCreateSecret = jest
      .fn()
      .mockImplementation(() => Promise.resolve({} as CreateSecretResponse));

    jest.spyOn(client, "CreateSecret").mockImplementation(mockClientCreateSecret);
    jest.spyOn(Secret, "resourcesClient").mockImplementation(() => client);
  });
  afterEach(() => {
    jest.restoreAllMocks();
  });

  it("creates a pull secret", async () => {
    const cluster = "default";
    const name = "repo-1";
    const user = "foo";
    const password = "pass";
    const email = "foo@bar.com";
    const server = "docker.io";
    const namespace = "default";

    await Secret.createPullSecret("default", name, user, password, email, server, namespace);

    expect(mockClientCreateSecret).toHaveBeenCalledWith({
      context: {
        cluster,
        namespace,
      },
      name,
      stringData: {
        ".dockerconfigjson":
          '{"auths":{"docker.io":{"username":"foo","password":"pass","email":"foo@bar.com","auth":"Zm9vOnBhc3M="}}}',
      },
      type: SecretType.SECRET_TYPE_DOCKER_CONFIG_JSON,
    } as CreateSecretRequest);
  });
});
