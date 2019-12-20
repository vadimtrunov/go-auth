FROM golang:1.13.5-alpine
RUN mkdir /app
WORKDIR /app

ADD ./app /app
RUN ls -al
RUN apk update; apk add git;
RUN go mod vendor && cd src/ && go build -o /app/bin .
CMD ["/app/bin"]