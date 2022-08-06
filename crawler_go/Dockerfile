FROM lehoanglong/go_playwright

WORKDIR /opt/app
COPY ./ /opt/app
RUN go build ./cmd/main.go && go install  ./cmd/main.go