# Copyright 2019-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

FROM mcr.microsoft.com/playwright:v1.23.0-focal
ADD . /app/
# Make sure we don't ship the current tests with the image
# Renaming or relocating tests might mess things up otherwise
RUN rm -rf /app/tests
RUN yarn install

# RUN npm install
RUN npx playwright install

# Run playwright test
CMD [ "npx", "playwright", "test", "--workers=1" ]
