FROM golang:1.22

ENV GITHUB_TOKEN=${GITHUB_TOKEN}

COPY . /home/src
WORKDIR /home/src
RUN go build -o /bin/cmd ./cmd

ENTRYPOINT [ "/bin/cmd" ]