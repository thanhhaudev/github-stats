FROM golang:1.22

COPY . /home/src
WORKDIR /home/src
RUN go build -o /bin/cmd ./cmd

ENTRYPOINT [ "/bin/cmd" ]