// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

#[test]
fn cli_tests() {
    trycmd::TestCases::new().case("tests/cmd/*.trycmd");
}
