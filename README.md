### exemplar
Golang microservice boilerplate


## Intro
Model repo to be used as template for creating microservices in golang


### Using this boilerplate
- Visit https://github.com/LambdatestIncPrivate/exemplar
- Click `Use this template` to create a new repository
- Git clone your new repo
- After cloning, add this repo to git remote of your project to regularly receive improvements.
  > Eg. `git remote add template https://github.com/LambdatestIncPrivate/exemplar`
- Then to update your repo with this app, fire `git pull template master`

## Building
- This can be built just like any other bo project `go build -o exemplar cmd/server.go`
- This project is docker ready. Dockerfile in the root directory can be used to create service docker
  - Sample docker usage
  ```
  docker build . --tag exemplar
  docker run -p 12223:12223  --env EXM_PORT=12223 exemplar
  ```
- A sample shell file is also included to automate the build steps. `build.sh` in the root directory can be used for special adding build steps.

### Docker compose
- Copy .env.example to .env
- Create GITHUB token from your account and replace in .env
- Run `docker-compose up` to setup all services
It exposes exemplar service HTTP over 9876 and GRPC over 12000 on host network.


## GRPC
This server exposes GRPC server on port 12000

### Grpcurl support
[Grpcurl](https://github.com/fullstorydev/grpcurl) is command line tool to debug and play with GRPC servers.
- List available services `grpcurl -plaintext localhost:12000 list`
- List available methods in services `grpcurl -plaintext localhost:12000 list bookkeeping.host.v1.HostService `
- Describe request `grpcurl -plaintext localhost:12000 describe  bookkeeping.host.v1.HostService.Get `
- Send create host grpc call
```
grpcurl -d @ -plaintext localhost:12000 bookkeeping.host.v1.HostService/Set <<EOM
{"host": {"privateIp":"10.0.0.1", "publicIp": "253.12.23.11"}}
EOM
```
- Send Get host grpc call
```
grpcurl -d @ -plaintext localhost:12000 bookkeeping.host.v1.HostService/Get <<EOM
{"uuid": "17820b00-d436-4b7b-b30d-cc46051ddc1a"}
EOM
```
- Send Set call using grpc_cli
```
grpc_cli call localhost:12000 bookkeeping.host.v1.HostService/Set --json_input '{"host": {"privateIp":"10.0.0.9", "publicIp": "122.122.255.255"}}' -metadata "x-key:sdfsd:x-key:someother:oh-yeah:okj"
```

## Migrations
For migrations we are using [go-migrate](https://github.com/golang-migrate/migrate). Download the binary in your system to create and apply migrations
For mac you can use `brew install golang-migrate` to install migrate binary.
Currently, migrations are stored inside `db/migrations` directory.
> Assuming you have a MySQL 8 running on your system with exemplar database
### Workflow
- Apply migrations `migrate -path db/migration -database "mysql://admin:password@(127.0.0.1:3306)/exemplar" -verbose up`
- Create new migration `migrate create -ext sql -dir db/migration -seq init_schema `. You need to populate both up and down sql files
For more details visit [this link](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

### Hot reloading
This project is configured to support [fresh](https://github.com/gravityblast/fresh) runner which reloads the application actively whenever any golang file (or any other file configured for hot reloading) changes. This is very useful while actively developing as it removes the need to recompile and run the application again and again. `runner.conf` in the root directory is used to configure the fresh runner. More information can be viewed on [their github project](https://github.com/gravityblast/fresh)

## Libraries used
- Grpc for transport
- [sqlx](https://github.com/jmoiron/sqlx/)
- [Goqu for db query creaton](https://github.com/doug-martin/goqu)
- [Viper](github.com/spf13/viper)
- [Cobra](github.com/spf13/cobra)
- [Backoff](github.com/cenkalti/backoff)
- [Go-kit](github.com/go-kit/kit)

## TODO
* [X] Add grpc example
* [X] Add go-kit integration
* [X] Add Migration
* [X] Add sqlx models
* [X] Add repository pattern
* [ ] Add params in context for grpc
* [X] Enable grpc refelction
* [X] Add base crud struct for embedding
* [X] Add go-kit compatible http transport using gin
* [ ] Add grpc client with retry
* [ ] Add unit test cases
* [ ] Add integration tests
* [ ] Add opencensus tracing
* [ ] Add health checks
* [ ] Add components and structure from modern go application repo
* [ ] Add swagger definitions for HTTP handlers
* [ ] Try emperror for error handling
* [ ] Add authentication over GRPC
* [ ] Add CI/CD using actions


## Libraries for future usecases
- Validaations: https://github.com/go-ozzo/ozzo-validation/

## Reading
- https://shijuvar.medium.com/go-microservices-with-go-kit-introduction-43a757398183
- https://sagikazarmark.hu/blog/getting-started-with-go-kit/
- https://threedots.tech/post/repository-pattern-in-go/
- https://threedots.tech/post/introducing-clean-architecture/
- https://threedots.tech/post/ddd-lite-in-go-introduction/?utm_source=about-wild-workouts#thats-great-but-do-you-have-any-evidence-if-that-is-working
- https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md

