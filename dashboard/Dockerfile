FROM node:8.9 AS build
WORKDIR /app

COPY package.json yarn.lock /app/
RUN yarn install --frozen-lockfile

COPY . /app
RUN yarn run build

FROM bitnami/nginx:1.14.0-r27
COPY --from=build /app/build /app
