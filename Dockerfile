# -------- STAGE 1: Build --------
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build for AWS Lambda (linux + amd64)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main

# -------- STAGE 2: Runtime --------
FROM public.ecr.aws/lambda/go:1

# Copy binary into Lambda task root
COPY --from=builder /app/main ${LAMBDA_TASK_ROOT}

# Command for Lambda
CMD ["main"]