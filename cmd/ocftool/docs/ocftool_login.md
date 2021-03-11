## ocftool login

Log in to a Gateway server

```
ocftool login [OPTIONS] [SERVER] [flags]
```

### Examples

```
# start interactive setup
ocftool login

# specify server name and user 
ocftool login localhost:8080 -u user

```

### Options

```
  -h, --help              help for login
  -p, --password string   Password
  -u, --username string   Username
```

### SEE ALSO

* [ocftool](ocftool.md)	 - CLI tool for working with OCF manifest files
