## gomod

> go mod manager: analyzed and upgrade project dependencies

### install

```shell
go install gitter.top/apps/gomod/...@latest
```



### usage

you can upgrade go.mod using:

```shell
x@fedora:~/GolandProjects/gomod$ gomod u
[INFO] [url=github.com/google/go-github/v64@v64.0.0] upgrade success
[INFO] [url=github.com/sirupsen/logrus@v1.9.3] upgrade success
[INFO] [url=github.com/spf13/cobra@v1.8.1] upgrade success
[INFO] [url=github.com/stretchr/testify@v1.9.0] upgrade success
[INFO] [url=gitter.top/common/lormatter@v0.0.1] upgrade success
[INFO] [url=golang.org/x/mod@v0.20.0] upgrade success
[INFO] [url=golang.org/x/net@v0.28.0] upgrade success
```

### feature

- [ ] `gomod` show go.mod available updates
- [x] `gomod u` upgrade go.mod
- [ ] `gomod a` analyzed project dependencies