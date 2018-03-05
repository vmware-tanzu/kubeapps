# Required k8s.io/helm to be checked out in your gopath
src/shared/hapi/release.js:
	./node_modules/.bin/pbjs -t static-module -w commonjs -o src/shared/hapi/release.js -p $$GOPATH/src/k8s.io/helm/_proto $$GOPATH/src/k8s.io/helm/_proto/hapi/release/release.proto && \
	./node_modules/.bin/pbts -o src/shared/hapi/release.d.ts src/shared/hapi/release.js && \
	echo "// tslint:disable\n$$(cat src/shared/hapi/release.js)" > src/shared/hapi/release.js && \
	echo "// tslint:disable\n$$(cat src/shared/hapi/release.d.ts)" > src/shared/hapi/release.d.ts
