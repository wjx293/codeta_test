# TemplateAdminFrontend - 多阶段构建（node builder → nginx runner）
FROM node:20-alpine AS builder

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .

# 空字符串表示同源请求，经 nginx 反代到 BFF（Cookie 会话需同源）
ARG VITE_API_BASE_URL=
ARG VITE_AUTH_MODE=wps
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_AUTH_MODE=$VITE_AUTH_MODE

RUN npm run build

FROM nginx:1.27-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 3001

CMD ["nginx", "-g", "daemon off;"]
