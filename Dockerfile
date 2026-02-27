# Stage 1: Build React frontend
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.23-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist cmd/server/static
RUN CGO_ENABLED=0 go build -o temple-ats ./cmd/server

# Stage 3: Final image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend /app/temple-ats .
COPY migrations/ migrations/
RUN mkdir -p uploads
EXPOSE 8080
CMD ["./temple-ats"]
