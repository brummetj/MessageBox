# Production use docker image

ARG BASE_IMAGE=debian:buster
FROM golang:1.15-buster AS builder
ENV APP_HOME /usr/local/mb/src
ENV MB_USER mb

RUN mkdir -p $APP_HOME && adduser $MB_USER \
    && chown $MB_USER:$MB_USER ${APP_HOME}

WORKDIR $APP_HOME
ADD . ../src

RUN go mod download && go mod verify && go build -o /usr/local/bin/mb ./cmd/http/main.go

FROM $BASE_IMAGE
ENV APP_HOME /usr/local/mb/src
ENV MB_USER mb

RUN mkdir -p $APP_HOME && adduser $MB_USER \
    && chown -R $MB_USER:$MB_USER $APP_HOME

WORKDIR $APP_HOME

COPY --from=builder $APP_HOME $APP_HOME
COPY --from=builder /usr/local/bin/mb /usr/local/bin/mb/

# Chown on COPY was giving me issues for whatever reason :(
RUN chown $MB_USER:$MB_USER $APP_HOME && chown $MB_USER:$MB_USER /usr/local/bin/mb /usr/local/bin/mb/
ENV PATH=$PATH:/usr/local/bin/mb
USER $MB_USER

CMD ["mb"]