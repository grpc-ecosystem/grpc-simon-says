# Java Client

This is a command line Java Client for our multiplayer gRPC game of [Simon](https://en.wikipedia.org/wiki/Simon_\(game\))!

## Compile
    $ mvn compile

## Run
    $ mvn package
    $ java -jar target/simonsays-1.0-SNAPSHOT.jar

There is also a Docker image, if you want to run it inside Docker

## Create Docker image
    $ docker build -t simonsays-java .

## Run Docker image
    $ docker run -it simonsays-java

## Licence
Apache 2.0

This is not an official Google Product.
