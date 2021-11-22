# Dependencies taken from https://github.com/puppeteer/puppeteer/blob/main/docs/troubleshooting.md#chrome-headless-doesnt-launch-on-unix
FROM bitnami/node:16.13.0
RUN install_packages ca-certificates fonts-liberation gconf-service libappindicator1 libappindicator3-1 libasound2 libatk1.0-0 libatk-bridge2.0-0 libc6 libcairo2 libcups2 libdbus-1-3 libdrm2 libexpat1 libfontconfig1 libgbm1 libgbm-dev libgcc1 libgconf-2-4 libgdk-pixbuf2.0-0 libglib2.0-0 libgtk-3-0 libice6 libnspr4 libnss3 libpango-1.0-0 libpangocairo-1.0-0 libsm6 libstdc++6 libx11-6 libx11-xcb1 libxcb1 libxcomposite1 libxcursor1 libxdamage1 libxext6 libxfixes3 libxi6 libxrandr2 libxrender1 libxss1 libxtst6 lsb-release wget xdg-utils libxshmfence-dev
ADD . /app/
RUN yarn install
CMD [ "yarn", "start" ]
