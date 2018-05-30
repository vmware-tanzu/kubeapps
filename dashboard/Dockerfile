FROM node:8.9 AS build
WORKDIR /app

COPY package.json yarn.lock /app/
RUN yarn install --frozen-lockfile

COPY . /app
RUN yarn run build

FROM bitnami/nginx
COPY --from=build /app/build /app
