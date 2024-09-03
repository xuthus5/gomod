## gomod

> go mod manager: analyzed and upgrade project dependencies

### install

```shell
go install gitter.top/apps/gomod/...@latest
```

### usage

you can show go.mod available updates using `gomod`:

```shell
x@fedora:~/GolandProjects/gomod$ gomod
             PACKAGE            | RELATION |              CURRENT               |               LATEST                
--------------------------------+----------+------------------------------------+-------------------------------------
  github.com/fatih/color        | direct   | v1.7.0                             | v1.17.0
  github.com/mattn/go-colorable | indirect | v0.1.2                             | v0.1.13
  github.com/mattn/go-isatty    | indirect | v0.0.8                             | v0.0.20
  github.com/mattn/go-runewidth | indirect | v0.0.9                             | v0.0.16
  golang.org/x/tools            | indirect | v0.13.0                            | v0.24.0
  golang.org/x/xerrors          | indirect | v0.0.0-20191204190536-9bdfabe68543 | v0.0.0-20240903120638-7835f813f4da
  gopkg.in/check.v1             | indirect | v0.0.0-20161208181325-20d25e280405 | v1.0.0-20201130134442-10cb98267c6c
```

upgrade go.mod using `gomod u`:

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

analyzed project dependencies:

```shell
x@fedora:~/GolandProjects/gomod$ gomod u
                PACKAGE                | RELATION |              VERSION               | GOVERSION | TOOLCHAINS  
---------------------------------------+----------+------------------------------------+-----------+-------------
  gitter.top/apps/gomod                | main     |                                    |      1.20 |
  github.com/briandowns/spinner        | direct   | v1.23.1                            |      1.17 |
  github.com/cpuguy83/go-md2man/v2     | indirect | v2.0.4                             |      1.11 |
  github.com/davecgh/go-spew           | indirect | v1.1.1                             |           |
  github.com/fatih/color               | direct   | v1.7.0                             |           |
  github.com/google/go-cmp             | indirect | v0.6.0                             |           |
  github.com/google/go-github/v64      | direct   | v64.0.0                            |      1.21 |
  github.com/google/go-querystring     | indirect | v1.1.0                             |      1.10 |
  github.com/inconshreveable/mousetrap | indirect | v1.1.0                             |      1.18 |
  github.com/mattn/go-colorable        | indirect | v0.1.2                             |           |
  github.com/mattn/go-isatty           | indirect | v0.0.8                             |           |
  github.com/mattn/go-runewidth        | indirect | v0.0.9                             |       1.9 |
  github.com/olekukonko/tablewriter    | direct   | v0.0.5                             |      1.12 |
  github.com/pmezard/go-difflib        | indirect | v1.0.0                             |           |
  github.com/russross/blackfriday/v2   | indirect | v2.1.0                             |           |
  github.com/sirupsen/logrus           | direct   | v1.9.3                             |      1.13 |
  github.com/spf13/cobra               | direct   | v1.8.1                             |      1.15 |
  github.com/spf13/pflag               | indirect | v1.0.5                             |      1.12 |
  github.com/stretchr/objx             | indirect | v0.5.2                             |           |
  github.com/stretchr/testify          | direct   | v1.9.0                             |      1.17 |
  gitter.top/common/lormatter          | direct   | v0.0.1                             |      1.21 |
  golang.org/x/crypto                  | indirect | v0.26.0                            |           |
  golang.org/x/mod                     | direct   | v0.20.0                            |      1.18 |
  golang.org/x/net                     | direct   | v0.28.0                            |      1.18 |
  golang.org/x/sys                     | indirect | v0.24.0                            |      1.18 |
  golang.org/x/term                    | indirect | v0.23.0                            |      1.18 |
  golang.org/x/text                    | indirect | v0.17.0                            |           |
  golang.org/x/tools                   | indirect | v0.13.0                            |           |
  golang.org/x/xerrors                 | indirect | v0.0.0-20191204190536-9bdfabe68543 |      1.11 |
  gopkg.in/check.v1                    | indirect | v0.0.0-20161208181325-20d25e280405 |           |
  gopkg.in/yaml.v3                     | indirect | v3.0.1                             |           |
```

### feature

- [x] `gomod` show go.mod available updates
- [x] `gomod u` upgrade go.mod
- [x] `gomod a` analyzed project dependencies