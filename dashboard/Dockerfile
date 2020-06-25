FROM node:8.9 AS build
WORKDIR /app

COPY package.json yarn.lock /app/
RUN yarn install --frozen-lockfile

RUN mkdir /app/src
RUN curl -L https://unpkg.com/@bitnami/hex@3.2.0/dist/hex.min.css > /app/src/hex.min.css
RUN curl -L https://unpkg.com/@clr/ui@3.1.4/clr-ui.min.css > /app/src/clr-ui.min.css
COPY . /app
RUN yarn run prettier-check
RUN yarn run build

FROM bitnami/nginx:1.18.0-debian-10-r38
COPY --from=build /app/build /app
