FROM golang:1.19

# INSTALL DEPENDENCIES
COPY go.mod go.sum /src/
RUN cd /src && go mod download

# BUILD BINARY
COPY . /src
RUN cd /src && make build

# DEFAULT DATA DIR
ENV DATA=/data
# DEFAULT FILE NAME
ENV URLS=urls.txt

# COPY BINARY INTO $PATH
COPY /dist/craurl /usr/local/bin/

CMD ["/bin/sh", "-c", "craurl $DATA/$URLS"]

