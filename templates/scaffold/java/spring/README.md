# {{.DBNameCamel}} Project

## Prerequisites

This project requires [Docker](https://docs.docker.com/get-docker/) and [Maven](https://maven.apache.org/install.html) to be installed.

## Starting Project Locally

```sh
mvn clean compile exec:java -Dexec.mainClass="{{.PackageName}}.{{.DBNameCamel}}Application"
```

This will start up the project at http://localhost:8080

To start the project in the detached docker container run:

```sh
docker compose --build -d
```

## Project Structure

├── docker-compose.yml
├── Dockerfile
├── pom.xml
├── README.md
└── src
└── main
├── java
│   └── {{.DBName}}
│       ├── collections
{{- $len := (add (len .Collections) -1)}}
{{- range $k, $v := .Collections}}
{{- if gt $len $k}}
│       │   ├── {{$v.Name}}.java
{{- else}}
│       │   └── {{$v.Name}}.java
{{- end}}
{{- end}}
│       ├── controller
{{- $len := (add (len .Collections) -1)}}
{{- range $k, $v := .Collections}}
{{- if gt $len $k}}
│       │   ├── {{$v.Name}}Controller.java
{{- else}}
│       │   └── {{$v.Name}}Controller.java
{{- end}}
{{- end}}
│       ├── {{.DBName}}Application.java
│       └── spring
│           ├── TigrisInitializer.java
│           └── TigrisSpringConfiguration.java
└── resources
├── application.yml
└── logback.xml

### Data Models
The `src/main/java/{{.DBName}}/collections` directory contains collections models, which is basically the structure of the document persisted
in the particular collection.
{{if gt (len .Collections) 0}}
For example:

```java
{{- (index .Collections 0).Schema -}}
```
{{- end}}

This model types can be modified to add new fields to the document.

### Routes

The `src/main/java/{{.DBName}}/controllers` directory contains classes which setup [Spring](https://spring.io) Web framework CRUD routes for every collection model.
Once project is started, they can be tested using curl commands.
{{if gt (len .Collections) 0}}
For example:
{{with (index .Collections 0)}}
#### Create document in the `{{.JSON}}` collection:
```
curl -X POST "localhost:8080/{{.JSON}}" -H 'Content-Type: application/json' 
    -d "{ JSON document body corresponding to the collections.{{.Name}} }"
```

#### Read document from the `{{.JSON}}` collection:
```
curl -X GET "localhost:8080/{{.JSON}}/{document id}"
```

#### Delete document from the `{{.JSON}}` collection:
```
curl -X DELETE "localhost:8080/{{.JSON}}/{document id}"
```
{{end}}
{{- end}}
Full Tigris documentation [here](https://docs.tigrisdata.com).

Be brave. Have fun!
