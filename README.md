# config validator
Validation tool for yaml configuration  

# install:

go get github.com/sklevenz/config-validator

# test:

```
go test
```

# build & run sample:

```
go run ConfigValidator.go help show
go run ConfigValidator.go help validate

go run ./ConfigValidator.go show examples/schema.yml

go run ./ConfigValidator.go validate examples/schema.yml examples/config.yml examples/credentials.yml
```

# help

```
usage: ConfigValidator [<flags>] <command> [<args> ...]

A validation tool for yaml configuration

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --version  Show application version.

Commands:
  help [<command>...]
    Show help.

  show [<flags>] <schema>
    Show schema definition

  validate <schema> <config>...
    Validate schema definition
```

## help - show

```
usage: ConfigValidator show [<flags>] <schema>

Show schema definition

Flags:
  --help               Show context-sensitive help (also try --help-long and --help-man).
  --version            Show application version.

Args:
  <schema>  Absolute file name to a schema yaml file
```
## help - validate

```
usage: ConfigValidator validate <schema> <config>...

Validate schema definition

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --version  Show application version.

Args:
  <schema>  Absolute file name to a schema yaml file
  <config>  Absolute file names to a configuration yaml files

```

# Schema

```
---
schema:					# schema definition
  group: <name>			# properties group
  - property: <name>	# e.g. config.metrics.retention_policy
    annotations:
    - required
    description: <string>
    default: <value>
```