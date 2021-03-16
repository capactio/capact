## ocftool

Collective Capability Manager CLI

### Synopsis

            _
  _    _  _|_  _|_   _    _   |
 (_)  (_   |    |_  (_)  (_)  |

ocftool - Collective Capability Manager CLI

A utility for managing Project Voltron & assist with authoring OCF content

To begin working with Project Voltron using the ocftool CLI, start with:

    $ ocftool login

NOTE: If you would like to use 'pass' for credential storage, be sure to
      set CAPECTL_CREDENTIALS_STORE_BACKEND to 'pass' in your shell's env variables.

      In order to watch follow the progress of the workflow execution, it is required
      to have 'kubectl' configured with the default context set to the same cluster where
      the Gateway URL points to.

Quick Start:

    $ ocftool hub interfaces search                    # Lists available content (generic interfaces)
    $ ocftool hub interfaces browse                    # Interactively browse available content in your terminal
    $ ocftool action search                            # Lists available actions in the 'default' namespace
    $ ocftool action get <action name> -n <namespace>  # Gets the details of a specified action in the specified namespace (table format)
    $ ocftool action get <action name> -o json         # Gets the details of a specified action in the 'default' namespace (JSON format)
    $ ocftool action run <action name>                 # Accepts the rendered action, and sends it to the workflow engine
    $ ocftool action status @latest                    # Gets the status of the last triggered action
    $ ocftool action watch <action name>               # Watches the workflow engine's progress while processing the specified action

    

```
ocftool [flags]
```

### Options

```
  -h, --help   help for ocftool
```

### SEE ALSO

* [ocftool action](ocftool_action.md)	 - This command consists of multiple subcommands to interact with target Actions
* [ocftool config](ocftool_config.md)	 - Manage configuration
* [ocftool hub](ocftool_hub.md)	 - This command consists of multiple subcommands to interact with Hub server.
* [ocftool login](ocftool_login.md)	 - Login to a Hub (Gateway) server
* [ocftool logout](ocftool_logout.md)	 - Logout from the Hub (Gateway) server
* [ocftool validate](ocftool_validate.md)	 - Validate OCF manifests

