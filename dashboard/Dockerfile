FROM bitnami/node:12.18.3 AS build
WORKDIR /app

COPY package.json yarn.lock /app/
RUN yarn install --frozen-lockfile

RUN mkdir /app/src
RUN curl -L https://unpkg.com/@clr/ui@4.0.1/clr-ui.min.css > /app/node_modules/@clr/ui/clr-ui.min.css
COPY . /app
RUN yarn run prettier-check && yarn run ts-compile-check
RUN yarn run build

FROM bitnami/nginx:1.19.2-debian-10-r32
COPY --from=build /app/build /app
