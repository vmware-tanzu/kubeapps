Using cobra for commands. Currently only root command to run server

Requires go 1.16 (due to MapFS being used for testing).

```bash
make cli-dependencies
```

Ensure GOPATH is on path.

Using `buf` to handle grpc generation etc. Need to install.

TODO:

* Add second plugin (separating packages from packagerepositories) API points.
* Enforce required methods for each api.
* Add required authz (will need for actual calls anyway).
* Actually serve generated swagger files
* Other cmds to demo
