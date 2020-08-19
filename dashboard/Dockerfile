FROM bitnami/node:12.18.3 AS build
WORKDIR /app

COPY package.json yarn.lock /app/
RUN yarn install --frozen-lockfile

RUN mkdir /app/src
RUN curl -L https://unpkg.com/@bitnami/hex@3.2.0/dist/hex.min.css > /app/src/hex.min.css
RUN curl -L https://unpkg.com/@clr/ui@3.1.4/clr-ui.min.css > /app/src/clr-ui.min.css
COPY . /app
RUN yarn run prettier-check && yarn run ts-compile-check
RUN yarn run build

FROM bitnami/nginx:1.19.2-debian-10-r1
COPY --from=build /app/build /app
