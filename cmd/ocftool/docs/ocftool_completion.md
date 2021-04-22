## ocftool completion

Generate shell completion scripts

### Synopsis

Generate shell completion scripts for Capact CLI commands.

If you need to set up completions manually, follow the instructions below. The exact
config file locations might vary based on your system. Make sure to restart your
shell before testing whether completions are working.

### bash
  Run this command:
  	echo "source <(argo completion bash)" >> ~/.bashrc

### zsh
  Generate a _ocftool completion script and put it somewhere in your $fpath:
  	ocftool completion zsh > /usr/local/share/zsh/site-functions/_ocftool
  
  Ensure that the following is present in your ~/.zshrc:
  	autoload -U compinit
  	compinit -i

  Zsh version 5.7 or later is recommended.

### fish
  Generate a ocftool.fish completion script:
  	ocftool completion fish > ~/.config/fish/completions/ocftool.fish


```
ocftool completion [bash|zsh|fish|powershell] [flags]
```

### Options

```
  -h, --help   help for completion
```

### SEE ALSO

* [ocftool](ocftool.md)	 - Collective Capability Manager CLI

