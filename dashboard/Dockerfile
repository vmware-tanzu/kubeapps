FROM node:8.9 AS build
WORKDIR /app

COPY package.json yarn.lock /app/
RUN yarn install --frozen-lockfile

RUN mkdir /app/src
COPY . /app
RUN yarn run prettier-check
RUN curl -L https://unpkg.com/@bitnami/hex@3.2.0/dist/hex.min.css > /app/src/hex.min.css
RUN yarn run build

FROM bitnami/nginx:1.18.0-debian-10-r38
COPY --from=build /app/build /app
