FROM bitnami/node:8

# Install yarn
RUN install_packages apt-transport-https && \
  curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - && \
  echo "deb https://dl.yarnpkg.com/debian/ stable main" >> /etc/apt/sources.list.d/yarn.list && \
  install_packages yarn

RUN yarn global add @angular/cli@1.4.3 && ng set --global packageManager=yarn

COPY rootfs /

EXPOSE 4200 49152

ENTRYPOINT ["/app-entrypoint.sh"]

CMD ["ng", "serve", "--host", "0.0.0.0"]
