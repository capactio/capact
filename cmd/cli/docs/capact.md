---
title: capact
---

## capact

Collective Capability Manager CLI

### Synopsis

```

  _   _.  ._    _.   _  _|_
 (_  (_|  |_)  (_|  (_   |_
          |

```

capact - Collective Capability Manager CLI

A utility that manages Capact resources and assists with creating OCF content.

To begin working with Capact using the capact CLI, start with:

    $ capact login

NOTE: If you would like to use 'pass' for credential storage, be sure to
      set CAPACT_CREDENTIALS_STORE_BACKEND to 'pass' in your shell's env variables.

      In order to watch follow the progress of the workflow execution, it is required
      to have 'kubectl' configured with the default context set to the same cluster where
      the Gateway URL points to.

Quick Start:

    $ capact hub interfaces search                    # Lists available content (generic interfaces)
    $ capact hub interfaces browse                    # Interactively browse available content in your terminal
    $ capact action search                            # Lists available actions in the 'default' namespace
    $ capact action get <action name> -n <namespace>  # Gets the details of a specified action in the specified namespace (table format)
    $ capact action get <action name> -o json         # Gets the details of a specified action in the 'default' namespace (JSON format)
    $ capact action run <action name>                 # Accepts the rendered action, and sends it to the workflow engine
    $ capact action status @latest                    # Gets the status of the last triggered action
    $ capact action watch <action name>               # Watches the workflow engine's progress while processing the specified action

    

```
capact [flags]
```

### Options

```
  -c, --config string                 Path to the YAML config file
  -h, --help                          help for capact
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - tracing (default 0 - disable)
```

### SEE ALSO

* [capact action](capact_action.md)	 - This command consists of multiple subcommands to interact with target Actions
* [capact alpha](capact_alpha.md)	 - Alpha features
* [capact completion](capact_completion.md)	 - Generate shell completion scripts
* [capact config](capact_config.md)	 - Manage configuration
* [capact environment](capact_environment.md)	 - This command consists of multiple subcommands to interact with a Capact environments
* [capact hub](capact_hub.md)	 - This command consists of multiple subcommands to interact with Hub server.
* [capact install](capact_install.md)	 - install Capact into a given environment
* [capact login](capact_login.md)	 - Login to a Hub (Gateway) server
* [capact logout](capact_logout.md)	 - Logout from the Hub (Gateway) server
* [capact manifest](capact_manifest.md)	 - This command consists of multiple subcommands to interact with OCF manifests
* [capact policy](capact_policy.md)	 - This command consists of multiple subcommands to interact with Policy
* [capact typeinstance](capact_typeinstance.md)	 - This command consists of multiple subcommands to interact with target TypeInstances
* [capact upgrade](capact_upgrade.md)	 - Upgrades Capact
* [capact version](capact_version.md)	 - Show version information about this binary

