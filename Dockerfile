FROM golang:1.12 as builder

WORKDIR /go/src/github.com/radu-matei/azure-functions-golang-worker
COPY . .

# RUN go get -u github.com/golang/dep/...
# RUN dep ensure

RUN go build -o golang-worker

# compile HTTP Trigger functions that works without any Azure account
RUN go build -buildmode=plugin -o functions/v2/bin/v2.so functions/v2/main.go

FROM mcr.microsoft.com/azure-functions/base:2.0

COPY workers/go /azure-functions-host/workers/go
COPY --from=builder /go/src/github.com/radu-matei/azure-functions-golang-worker/golang-worker /azure-functions-host/workers/go/golang-worker
RUN chmod +x /azure-functions-host/workers/go/golang-worker

COPY --from=builder /go/src/github.com/radu-matei/azure-functions-golang-worker/functions/ /home/site/wwwroot

ENV AzureWebJobsScriptRoot=/home/site/wwwroot \
    HOME=/home \
    ASPNETCORE_URLS=http://+:80 \
    AZURE_FUNCTIONS_ENVIRONMENT=Development
