package cmd

import (
	"fmt"
	"os"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/heredoc"

	"github.com/spf13/cobra"
)

func NewCompletion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: heredoc.WithCLIName(`
			Generate shell completion scripts for <cli> CLI commands.
			
			If you need to set up completions manually, follow the instructions below. The exact
			config file locations might vary based on your system. Make sure to restart your
			shell before testing whether completions are working.
			
			### bash
			  Run this command:
			  	echo "source <(<cli> completion bash)" >> ~/.bashrc
			
			### zsh
			  Generate a _<cli> completion script and put it somewhere in your $fpath:
			  	<cli> completion zsh > /usr/local/share/zsh/site-functions/_<cli>
			  
			  Ensure that the following is present in your ~/.zshrc:
			  	autoload -U compinit
			  	compinit -i
			
			  Zsh version 5.7 or later is recommended.
			
			### fish
			  Generate a <cli>.fish completion script:
			  	<cli> completion fish > ~/.config/fish/completions/<cli>.fish
		`, cli.Name),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				shellType = args[0]
				w         = os.Stdout
				rootCmd   = cmd.Parent()
			)

			switch shellType {
			case "bash":
				return rootCmd.GenBashCompletion(w)
			case "zsh":
				return rootCmd.GenZshCompletion(w)
			case "powershell":
				return rootCmd.GenPowerShellCompletion(w)
			case "fish":
				return rootCmd.GenFishCompletion(w, true)
			default:
				return fmt.Errorf("unsupported shell type %q", shellType)
			}
		},
	}

	return cmd
}
