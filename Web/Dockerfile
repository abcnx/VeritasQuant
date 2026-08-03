# FinvQuant 前端镜像（Vue3 + Vite8 + Vuetify4 → Nginx）
# 默认端口 16002，/api 反向代理到服务端 16001

# ---- 构建阶段 ----
FROM node:24-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

# ---- 运行阶段 ----
FROM nginx:alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 16002

CMD ["nginx", "-g", "daemon off;"]
