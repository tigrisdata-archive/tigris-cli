# {{.DBNameCamel}} Project

## Prerequisites

This project requires [Docker](https://docs.docker.com/get-docker/) and [NPM](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm) to be installed.

## Starting Project

```sh
npm ci
npm start
```

This will start up the project at http://localhost:3000 and connect to production Tigris instance.

Execute `npm run dev` to start application connected to local Tigris instance.

## Project Structure

```
├── docker-compose.yml
├── Dockerfile
├── package.json
├── package-lock.json
├── README.md
├── src
│   ├── routes
│   │   ├── routes.ts
{{- $len := (add (len .Collections) -1)}}
{{- range $k, $v := .Collections}}
{{- if gt $len $k}}
│   │   ├── {{$v.JSON}}.ts
{{- else}}
│   │   └── {{$v.JSON}}.ts
{{- end}}
{{- end}}
│   ├── index.ts
│   └── models
{{- $len := (add (len .Collections) -1)}}
{{- range $k, $v := .Collections}}
{{- if gt $len $k}}
│   │   ├── {{$v.JSON}}.ts
{{- else}}
│   │   └── {{$v.JSON}}.ts
{{- end}}
{{- end}}
└── tsconfig.json
```

### Data Models

The `src/models` directory contains collections models, which is basically the structure of the document persisted
in the particular collection.
{{if gt (len .Collections) 0}}
For example:

```typescript
{{- (index .Collections 0).Schema -}}
```
{{- end}}

This model types can be modified to add new fields to the document.

### Routes

The `src/routes` directory contains classes which setup [Express](https://expressjs.com) Web framework CRUD routes for every collection model.
Once project is started, they can be tested using curl commands.
{{if gt (len .Collections) 0}}
For example:
{{with (index .Collections 0)}}
#### Create document in the `{{.JSON}}` collection:
```
curl -X POST "localhost:3000/{{.JSON}}" -H 'Content-Type: application/json' 
    -d "{ JSON document body corresponding to the models.{{.Name}} }"
```

#### Read document from the `{{.JSON}}` collection:
```
curl -X GET "localhost:3000/{{.JSON}}/{document id}"
```

#### Delete document from the `{{.JSON}}` collection:
```
curl -X DELETE "localhost:3000/{{.JSON}}/{document id}"
```
{{end}}
{{- end}}
Full Tigris documentation [here](https://docs.tigrisdata.com).

Be brave. Have fun!
