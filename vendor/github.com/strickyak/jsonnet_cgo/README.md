jsonnet_cgo
===========

Simple golang cgo wrapper around JSonnet VM.

Everything in libjsonnet.h is covered except the multi-file evaluators.

See jsonnet_test.go for how to use it.

Quick example:

        vm := jsonnet.Make()
        vm.ExtVar("color", "purple")

        x, err := vm.EvaluateSnippet(`Test_Demo`, `"dark " + std.extVar("color")`)

        if err != nil {
                panic(err)
        }
        if x != "\"dark purple\"\n" {
                panic("fail: we got " + x)
        }

        vm.Destroy()

