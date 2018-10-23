FROM go-dep:latest

COPY . ./

WORKDIR $GOPATH/src/github.com/openlab-red/mutating-webhook-vault-agent
ENV HOME=/home/mutating-webhook-vault-agent

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o $HOME/app .

RUN chown -R 1001:0 $HOME && \
    chmod -R g+rw $HOME

WORKDIR $HOME

USER 1001

ENTRYPOINT ["./app"]
