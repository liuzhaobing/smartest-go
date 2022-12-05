FROM go:1.18.8

MAINTAINER liuzhaobing

ADD . /opt/smartest-go/

WORKDIR /opt/smartest-go/

RUN go mod download

RUN go mod verify

RUN go build smartest-go

ENV HOST 0.0.0.0

ENV PORT 27997

ENV PORT 27997

CMD ["smartest-go", "-d", "27997"]