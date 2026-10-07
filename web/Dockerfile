FROM --platform=$BUILDPLATFORM node:24-alpine AS build
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:1.30-alpine
# CI sets APK_REFRESH so a newly published Alpine fix is not hidden by the layer cache.
ARG APK_REFRESH=local
RUN apk upgrade --no-cache
COPY nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /src/dist /usr/share/nginx/html
EXPOSE 80 8081
