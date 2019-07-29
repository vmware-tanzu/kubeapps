# Creates a node-8 debian image with telepresence installed and relevant tooling.
FROM bitnami/node:8-debian-9

RUN apt-get update -y \
  # See Debian instructions at https://www.telepresence.io/reference/install#ubuntu-1604-or-later
  && curl -sO https://packagecloud.io/install/repositories/datawireio/telepresence/script.deb.sh \
  && env os=ubuntu dist=xenial bash script.deb.sh \
  && apt-get install -y --no-install-recommends telepresence \
  && rm script.deb.sh \
  # Not sure why iptables and sudo are not pulled in as deps of telepresence, perhaps due to the above
  # hack of pretending to be ubuntu (and telepresence not expecting to run in a docker container).
  && apt-get install -y iptables sudo \
  # kubectl
  && curl -sSL -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
  && chmod +x /usr/local/bin/kubectl \
  # Typescript deps
  && npm install -g tslint typescript \
  # Cleanup
  && apt-get autoremove -y \
  && apt-get clean -y \
  && rm -rf /var/lib/apt/lists/*

ENV PORT 3000

COPY rootfs/ /
ENTRYPOINT ["/entrypoint.sh"]

CMD ["yarn", "start"]
