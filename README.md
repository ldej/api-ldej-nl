# api-ldej-nl

## gcloud

```shell
$ gcloud auth login
$ gcloud init
$ gcloud --project=api-ldej-nl app deploy --quiet
```

## Swagger

```shell
$ go get github.com/go-swagger/go-swagger/cmd/swagger
$ go install github.com/go-swagger/go-swagger/cmd/swagger
```

To serve SwaggerUI:

```shell
$ git clone https://github.com/swagger-api/swagger-ui
$ cp -r swagger-ui/dist swagger
```

Update app.yaml:

```yaml
- url: /swagger
  static_dir: swagger
```
  
Add static file handler for local:

```go
router.Handle("/swagger/*", http.StripPrefix("/swagger", http.FileServer(http.Dir("swagger"))))
```

Update swagger docs:

```shell
$ swagger generate spec -o ./swagger/swagger.json --scan-models
```