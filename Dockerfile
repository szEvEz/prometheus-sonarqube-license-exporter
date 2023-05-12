FROM golang:1.19-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 go build -o /prometheus-sonarqube-license-exporter

FROM scratch

COPY --from=build /prometheus-sonarqube-license-exporter .

CMD ["/prometheus-sonarqube-license-exporter"]
