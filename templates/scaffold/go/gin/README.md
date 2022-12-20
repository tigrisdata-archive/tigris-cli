# {{.DBNameCamel}} Project

## Prerequisites

This project requires [Docker](https://docs.docker.com/get-docker/) and [Task](https://taskfile.dev/installation/) to be installed.

## Starting Project Locally

```sh
task run
```

This will start up the project at http://localhost:8080

Executing `task run:docker` will start the project in the detached docker container.

Run `task` without arguments to see all available commands.

## Project Structure

├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── main.go
├── model
{{- $len := (add (len .Collections) -1)}}
{{- range $k, $v := .Collections}}
{{- if gt $len $k}}
│   ├── {{$v.JSON}}.go
{{- else}}
│   └── {{$v.JSON}}.go
{{- end}}
{{- end}}
├── README.md
├── route
│   └── route.go
└── Taskfile.yaml

### Data Models
The `model` directory contains collections models, which is basically the structure of the document persisted
in the particular collection.
{{if gt (len .Collections) 0}}
For example:

```golang
{{- (index .Collections 0).Schema -}}
```
{{- end}}

This model types can be modified to add new fields to the document.

### Routes

The `route/routes.go` defines SetupCRUD function which is used in the `main.go` to set up [Gin](https://github.com/gin-gonic/gin)
Web framework CRUD routes for every collection model.
Once project is started, they can be tested using curl commands.
{{if gt (len .Collections) 0}}
For example:
{{with (index .Collections 0)}}
#### Create document in the `{{.JSON}}` collection:
```
curl -X POST "localhost:8080/{{.JSON}}" -H 'Content-Type: application/json' 
    -d "{ JSON document body corresponding to the model.{{.Name}} }"
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
