FROM maven:3.2-jdk-7-onbuild
ADD ./target/simonsays-1.0-SNAPSHOT.jar simonsays.jar
ENTRYPOINT ["java", "-jar", "simonsays.jar"]