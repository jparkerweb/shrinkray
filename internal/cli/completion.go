package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/jparkerweb/shrinkray/internal/presets"
)

// completionCmd returns the completion command with sub-commands for each shell.
func completionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for shrinkray.

To load completions:

Bash:
  source <(shrinkray completion bash)
  # Or add to ~/.bashrc:
  echo 'source <(shrinkray completion bash)' >> ~/.bashrc

Zsh:
  shrinkray completion zsh > "${fpath[1]}/_shrinkray"
  # Or for oh-my-zsh:
  shrinkray completion zsh > ~/.oh-my-zsh/completions/_shrinkray

Fish:
  shrinkray completion fish | source
  # Or persist:
  shrinkray completion fish > ~/.config/fish/completions/shrinkray.fish

PowerShell:
  shrinkray completion powershell | Out-String | Invoke-Expression
  # Or add to $PROFILE:
  shrinkray completion powershell >> $PROFILE
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}

	return cmd
}

// registerCompletions registers custom completion functions for flags.
func registerCompletions(rootCmd *cobra.Command) {
	// --preset completion: all preset keys
	_ = rootCmd.RegisterFlagCompletionFunc("preset", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		all := presets.All()
		keys := make([]string, 0, len(all))
		for _, p := range all {
			keys = append(keys, p.Key+"\t"+p.Name)
		}
		return keys, cobra.ShellCompDirectiveNoFileComp
	})

	// --codec completion
	_ = rootCmd.RegisterFlagCompletionFunc("codec", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"h264\tH.264/AVC",
			"h265\tH.265/HEVC",
			"av1\tAV1",
			"vp9\tVP9",
		}, cobra.ShellCompDirectiveNoFileComp
	})

	// --sort completion
	_ = rootCmd.RegisterFlagCompletionFunc("sort", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"size-asc\tSmallest first",
			"size-desc\tLargest first",
			"name\tAlphabetical",
			"duration\tShortest first",
		}, cobra.ShellCompDirectiveNoFileComp
	})

	// --log-level completion
	_ = rootCmd.RegisterFlagCompletionFunc("log-level", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"debug", "info", "warn", "error"}, cobra.ShellCompDirectiveNoFileComp
	})
}
