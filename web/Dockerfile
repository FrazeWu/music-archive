# Build stage
FROM --platform=linux/arm64 node:20-alpine AS build-stage

WORKDIR /app

# Copy package files
COPY web/package*.json ./
RUN npm ci

# Copy source code
COPY web/ ./

# Build the application
RUN npm run build

# Production stage
FROM --platform=linux/arm64 nginx:stable-alpine AS production-stage

# Copy built assets from build stage
COPY --from=build-stage /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY web/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]