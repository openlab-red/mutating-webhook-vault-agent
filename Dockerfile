FROM docker.io/golang:1.13
LABEL authors="Mattia Mascia <mmascia@redhat.com>"

WORKDIR $GOPATH/src/github.com/openlab-red/mutating-webhook-vault-agent

COPY . ./

ENV HOME=/opt/webhook

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o $HOME/app .

RUN chown -R 1001:0 $HOME && \
    chmod -R g+rw $HOME

WORKDIR $HOME

USER 1001

ENTRYPOINT ["./app"]
