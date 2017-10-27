## Managing vendor/

This project uses `govendor`.  To make things easier for reviewers,
put updates to `vendor/` in a separate commit within the same PR.

### Clean up removed/unnecessary dependencies

This is a good one to run before sending a non-trivial change for
review.

```
# See unused
govendor list +unused

# Remove unused
govendor remove +unused
```

### Add a new dependency

Make code change that imports new library, then:

```
# See missing
govendor list +missing

% govendor fetch -v +missing
```

Note pinning a library to a specific version requires extra work.  In
particular, we do this for `client-go` and `apimachinery` - to find
the current version used for these libraries, look in
`vendor/vendor.json` for the `version` field (not `versionExact`) of
an existing package.  Use that same version in the commands below:

```
# For example: To pin all imported client-go packages to release v3.0
% govendor fetch -v k8s.io/client-go/...@v3.0

# *Note* the above may pull in new packages from client-go, which will
# be imported at HEAD.  You need to re-run the above command until
# all imports are at the desired version.
# TODO: There is probably a better way to do this.
```

It is safe (and appropriate) to re-run `govendor fetch` with a
different version, if you made a mistake or missed some libraries.

## Making a Release

1. Add appropriate tag.  We do this via git (not github UI) so the tag
   is signed.  This process requires you to have write access to the
   real master branch (not your local fork).
   ```
   % tag=vX.Y.Z
   % git fetch   # update
   % git tag -s -m $tag $tag origin/master
   % git push origin tag $tag
   ```

2. Wait for the travis autobuilders to build release binaries.

3. *Now* create the github release, using the existing tag created
   above.
