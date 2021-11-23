# Contributing

When contributing to Kubeapps, please first discuss the change you wish to make via an issue with this repository before making a change.

> Kubeapps distribution is delegated to the official [Bitnami Kubeapps chart](https://github.com/bitnami/charts/tree/master/bitnami/kubeapps) from the separate Bitnami charts repository. PRs and issues related to the official chart should be created in the Bitnami charts repository.

Please note we have a [code of conduct](./CODE_OF_CONDUCT.md), please follow it in all your interactions with the project.

## Pull Request Process

1. Ensure any install or build dependencies are removed before the end of the layer when doing a build.
2. Update the README.md with details of changes to the interface, this includes new environment variables, exposed ports, useful file locations and container parameters.
3. Increase the version numbers in any examples files and the README.md to the new version that this Pull Request would represent. The versioning scheme we use is [SemVer](https://semver.org/).
4. You may merge the Pull Request in once you have the sign-off of two other developers, or if you do not have permission to do that, you may request the second reviewer to merge it for you.

## DCO Sign off

All authors to the project retain copyright to their work. However, to ensure
that they are only submitting work that they have rights to, we are requiring
everyone to acknowledge this by signing their work.

Any copyright notices in this repo should specify the authors as "the contributors".

To sign your work, just add a line like this at the end of your commit message:

```text
Signed-off-by: Michael Nelson <minelson@vmware.com>
```

This can easily be done with the `--signoff` option to `git commit`.

By doing this you state that you can certify the following (from [https://developercertificate.org/](https://developercertificate.org/):

```text
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

If you find yourself in the position where you've created a pull-request and it's failing the DCO check because some of your commits are not signed off, you can just follow the details of the DCO check failure to sign-off on those commits with a single command, such as:

```bash
git rebase HEAD~N --signoff
git push --force-with-lease
```

where N is the number of commits you've added.

You can also setup a [commit template for your local git config](https://stackoverflow.com/a/34687806) that includes your sign-off.
