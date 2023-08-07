// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// @generated by protoc-gen-connect-es v0.12.0 with parameter "target=ts,import_extension=none"
// @generated from file kubeappsapis/core/packages/v1alpha1/repositories.proto (package kubeappsapis.core.packages.v1alpha1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import {
  AddPackageRepositoryRequest,
  AddPackageRepositoryResponse,
  DeletePackageRepositoryRequest,
  DeletePackageRepositoryResponse,
  GetPackageRepositoryDetailRequest,
  GetPackageRepositoryDetailResponse,
  GetPackageRepositoryPermissionsRequest,
  GetPackageRepositoryPermissionsResponse,
  GetPackageRepositorySummariesRequest,
  GetPackageRepositorySummariesResponse,
  UpdatePackageRepositoryRequest,
  UpdatePackageRepositoryResponse,
} from "./repositories_pb";
import { MethodKind } from "@bufbuild/protobuf";

/**
 * Each repositories v1alpha1 plugin must implement at least the following rpcs:
 *
 *
 * @generated from service kubeappsapis.core.packages.v1alpha1.RepositoriesService
 */
export const RepositoriesService = {
  typeName: "kubeappsapis.core.packages.v1alpha1.RepositoriesService",
  methods: {
    /**
     * @generated from rpc kubeappsapis.core.packages.v1alpha1.RepositoriesService.AddPackageRepository
     */
    addPackageRepository: {
      name: "AddPackageRepository",
      I: AddPackageRepositoryRequest,
      O: AddPackageRepositoryResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kubeappsapis.core.packages.v1alpha1.RepositoriesService.GetPackageRepositoryDetail
     */
    getPackageRepositoryDetail: {
      name: "GetPackageRepositoryDetail",
      I: GetPackageRepositoryDetailRequest,
      O: GetPackageRepositoryDetailResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kubeappsapis.core.packages.v1alpha1.RepositoriesService.GetPackageRepositorySummaries
     */
    getPackageRepositorySummaries: {
      name: "GetPackageRepositorySummaries",
      I: GetPackageRepositorySummariesRequest,
      O: GetPackageRepositorySummariesResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kubeappsapis.core.packages.v1alpha1.RepositoriesService.UpdatePackageRepository
     */
    updatePackageRepository: {
      name: "UpdatePackageRepository",
      I: UpdatePackageRepositoryRequest,
      O: UpdatePackageRepositoryResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kubeappsapis.core.packages.v1alpha1.RepositoriesService.DeletePackageRepository
     */
    deletePackageRepository: {
      name: "DeletePackageRepository",
      I: DeletePackageRepositoryRequest,
      O: DeletePackageRepositoryResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kubeappsapis.core.packages.v1alpha1.RepositoriesService.GetPackageRepositoryPermissions
     */
    getPackageRepositoryPermissions: {
      name: "GetPackageRepositoryPermissions",
      I: GetPackageRepositoryPermissionsRequest,
      O: GetPackageRepositoryPermissionsResponse,
      kind: MethodKind.Unary,
    },
  },
} as const;
