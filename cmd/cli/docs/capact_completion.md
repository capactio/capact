---
title: capact completion
---

## capact completion

Generate shell completion scripts

### Synopsis

Generate shell completion scripts for capact CLI commands.

If you need to set up completions manually, follow the instructions below. The exact
config file locations might vary based on your system. Make sure to restart your
shell before testing whether completions are working.

### bash
  Run this command:
  	echo "source <(capact completion bash)" >> ~/.bashrc

### zsh
  Generate a _capact completion script and put it somewhere in your $fpath:
  	capact completion zsh > /usr/local/share/zsh/site-functions/_capact
  
  Ensure that the following is present in your ~/.zshrc:
  	autoload -U compinit
  	compinit -i

  Zsh version 5.7 or later is recommended.

### fish
  Generate a capact.fish completion script:
  	capact completion fish > ~/.config/fish/completions/capact.fish


```
capact completion [bash|zsh|fish|powershell] [flags]
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
  -C, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - trace (default 0 - disable)
```

### SEE ALSO

* [capact](capact.md)	 - Collective Capability Manager CLI

