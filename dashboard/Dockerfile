FROM node:8.9 AS build
WORKDIR /app

COPY package.json yarn.lock /app/
RUN yarn install --frozen-lockfile

RUN mkdir /app/src
RUN curl -L https://unpkg.com/@bitnami/hex@3.2.0/dist/hex.min.css > /app/src/hex.min.css

COPY . /app
RUN yarn run build

FROM bitnami/nginx:1.17.10-debian-10-r10
COPY --from=build /app/build /app
