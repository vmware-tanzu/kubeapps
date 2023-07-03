// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// @generated by protoc-gen-connect-es v0.11.0 with parameter "target=ts,import_extension=none"
// @generated from file kubeappsapis/plugins/resources/v1alpha1/resources.proto (package kubeappsapis.plugins.resources.v1alpha1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import {
  CanIRequest,
  CanIResponse,
  CheckNamespaceExistsRequest,
  CheckNamespaceExistsResponse,
  CreateNamespaceRequest,
  CreateNamespaceResponse,
  CreateSecretRequest,
  CreateSecretResponse,
  GetNamespaceNamesRequest,
  GetNamespaceNamesResponse,
  GetResourcesRequest,
  GetResourcesResponse,
  GetSecretNamesRequest,
  GetSecretNamesResponse,
  GetServiceAccountNamesRequest,
  GetServiceAccountNamesResponse,
} from "./resources_pb";
import { MethodKind } from "@bufbuild/protobuf";

/**
 * ResourcesService
 *
 * The Resources service is a plugin that enables some limited access to Kubernetes
 * resources on the cluster, using the user credentials sent with the request.
 *
 * @generated from service kubeappsapis.plugins.resources.v1alpha1.ResourcesService
 */
export const ResourcesService = {
  typeName: "kubeappsapis.plugins.resources.v1alpha1.ResourcesService",
  methods: {
    /**
     * @generated from rpc kubeappsapis.plugins.resources.v1alpha1.ResourcesService.GetResources
     */
    getResources: {
      name: "GetResources",
      I: GetResourcesRequest,
      O: GetResourcesResponse,
      kind: MethodKind.ServerStreaming,
    },
    /**
     * @generated from rpc kubeappsapis.plugins.resources.v1alpha1.ResourcesService.GetServiceAccountNames
     */
    getServiceAccountNames: {
      name: "GetServiceAccountNames",
      I: GetServiceAccountNamesRequest,
      O: GetServiceAccountNamesResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kubeappsapis.plugins.resources.v1alpha1.ResourcesService.GetNamespaceNames
     */
    getNamespaceNames: {
      name: "GetNamespaceNames",
      I: GetNamespaceNamesRequest,
      O: GetNamespaceNamesResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kubeappsapis.plugins.resources.v1alpha1.ResourcesService.CreateNamespace
     */
    createNamespace: {
      name: "CreateNamespace",
      I: CreateNamespaceRequest,
      O: CreateNamespaceResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kubeappsapis.plugins.resources.v1alpha1.ResourcesService.CheckNamespaceExists
     */
    checkNamespaceExists: {
      name: "CheckNamespaceExists",
      I: CheckNamespaceExistsRequest,
      O: CheckNamespaceExistsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kubeappsapis.plugins.resources.v1alpha1.ResourcesService.GetSecretNames
     */
    getSecretNames: {
      name: "GetSecretNames",
      I: GetSecretNamesRequest,
      O: GetSecretNamesResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kubeappsapis.plugins.resources.v1alpha1.ResourcesService.CreateSecret
     */
    createSecret: {
      name: "CreateSecret",
      I: CreateSecretRequest,
      O: CreateSecretResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kubeappsapis.plugins.resources.v1alpha1.ResourcesService.CanI
     */
    canI: {
      name: "CanI",
      I: CanIRequest,
      O: CanIResponse,
      kind: MethodKind.Unary,
    },
  },
} as const;
