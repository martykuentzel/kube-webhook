FROM golang:1.12-alpine AS build 
ENV GO111MODULE on

RUN apk add git

WORKDIR /go/src/github.com/MartyKuentzel/kube-webhook
COPY go.mod go.sum ./

RUN go mod download
COPY cmd cmd
COPY pkg pkg 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o sechook cmd/main.go

FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /go/src/github.com/MartyKuentzel/kube-webhook/sechook .
CMD ["/app/sechook"]


## what does cgo_enabled mean
## what does installsuffix do
## optimal docker bbuild
## what does ca-certificates mean in apk
## understand roberts docker file