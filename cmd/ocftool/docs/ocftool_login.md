## ocftool login

Login to a Hub (Gateway) server

```
ocftool login [OPTIONS] [SERVER] [flags]
```

### Examples

```
# start interactive setup
ocftool login

# Specify server name and specify the user
ocftool login localhost:8080 -u user

```

### Options

```
  -h, --help              help for login
  -p, --password string   Password
  -u, --username string   Username
```

### SEE ALSO

* [ocftool](ocftool.md)	 - Collective Capability Manager CLI

