# syntax=docker/dockerfile:1.7

FROM node:22-bookworm-slim AS builder
WORKDIR /web

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /web/dist /usr/share/nginx/html
RUN chown -R nginx:nginx /usr/share/nginx/html /var/cache/nginx /var/log/nginx /run
USER nginx
EXPOSE 8080
