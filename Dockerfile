FROM golang
 
ADD . /go/src/github.com/c3e/topicbot
RUN go install github.com/c3e/topicbot
ENTRYPOINT /go/bin/topicbot
