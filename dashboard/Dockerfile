FROM bitnami/node:12.20.0 AS build
WORKDIR /app

COPY package.json yarn.lock /app/
RUN yarn install --frozen-lockfile

RUN mkdir /app/src
COPY . /app
RUN yarn run prettier-check && yarn run ts-compile-check
RUN yarn run build

FROM bitnami/nginx:1.19.7-debian-10-r1
COPY --from=build /app/build /app
