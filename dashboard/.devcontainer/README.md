## Running the dev server in a container
To run the dev server in a container without telepresence or similar, run

```
docker-compose up
```
and then open a browser at http://localhost:3000/

## Running the test runner
You can run the frontend tests in a separate shell: 

```
docker-compose run dev-dashboard yarn test
```

## Running the dashboard via telepresence in-cluster
To instead run the local dev server via a telepresence shell, you can:
```
docker-compose run dev-dashboard bash
/telepresence-shell.sh
```
This will drop you into a telepresence shell, from which you can then `yarn start`. You can customise your KubeApps namespace and deployment with the env vars defined in the `docker-compose.yml`.

NOTE: I was unable so far to get telepresence working 100% correctly. The shell starts, but with the warnings:

```
T: Mounting remote volumes failed, they will be unavailable in this session. If you are running on Windows 
T: Subystem for Linux then see https://github.com/datawire/telepresence/issues/115, otherwise please report a 
T: bug, attaching telepresence.log to the bug report: https://github.com/datawire/telepresence/issues/new

T: Mount error was: fuse: mount failed: Permission denied
```
This is an issue running `fuse` / `sshfs` in a container, it seems. Need to check the HOWTO for Telepresence in containers mentioned in [175](https://github.com/telepresenceio/telepresence/issues/175)

Similarly, `yarn start` exits 1 without much more info when run from the telepresence shell. Not sure if it is related.

## Opening the container in VSCode
To open the dev container running in VSCode, assuming you have the [Remote Development Extension pack](https://code.visualstudio.com/docs/remote/remote-overview) installed, you can use the `Remote-Containers: Open Folder in Container` command to open the `kubeapps/dashboard` folder and VSCode will automatically start the container and attach itself, as well as stop the container when you close the remote session.

You can then use terminal windows within VSCode to run `yarn start` or `yarn test` or `/telepresence-shell.sh` as above.
