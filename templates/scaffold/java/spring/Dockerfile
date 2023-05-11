FROM maven:3.8-jdk-8 AS build
COPY . /app/
WORKDIR /app
RUN mvn clean package

FROM gcr.io/distroless/java11-debian11
COPY --from=build /app//target/*-exec.jar /app/main.jar
WORKDIR /app
CMD ["main.jar"]
