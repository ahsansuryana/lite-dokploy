# Stage 1: Build frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.23-alpine AS backend-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
COPY --from=frontend-builder /app/frontend/dist ./frontend-dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server

# Stage 3: Runtime
FROM alpine:3.20
RUN apk add --no-cache docker-cli docker-compose git ca-certificates tzdata
COPY --from=backend-builder /server /lite-dokploy
COPY --from=backend-builder /app/backend/frontend-dist /frontend

EXPOSE 3000
VOLUME ["/data", "/repositories", "/logs", "/deployments"]

ENTRYPOINT ["/lite-dokploy"]
CMD ["--port=3000", "--data=/data", "--repos=/repositories", "--logs=/logs", "--frontend=/frontend", "--traefik"]
