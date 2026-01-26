package cmd

import (
	"fmt"

	"github.com/jackchuka/gh-md/internal/discovery"
	"github.com/jackchuka/gh-md/internal/output"
	"github.com/spf13/cobra"
)

var reposFormat string

var reposCmd = &cobra.Command{
	Use:   "repos",
	Short: "List all managed repositories",
	Long: `List all repositories that have been pulled with gh-md.

Shows each repository with its last sync time.

Examples:
  gh md repos                  # List all repos
  gh md repos --format=json    # Output as JSON
  gh md repos --format=yaml    # Output as YAML`,
	RunE: runRepos,
}

func init() {
	rootCmd.AddCommand(reposCmd)

	reposCmd.Flags().StringVar(&reposFormat, "format", "text", "Output format (text, json, yaml)")
}

type repoOutput struct {
	Owner    string `json:"owner" yaml:"owner"`
	Repo     string `json:"repo" yaml:"repo"`
	Path     string `json:"path" yaml:"path"`
	LastSync string `json:"last_sync" yaml:"last_sync"`
}

func runRepos(cmd *cobra.Command, args []string) error {
	p := output.NewPrinter(cmd).WithFormat(output.ParseFormat(reposFormat))

	repos, err := discovery.DiscoverManagedRepos()
	if err != nil {
		return fmt.Errorf("failed to discover repositories: %w", err)
	}

	if len(repos) == 0 {
		p.Print("No managed repositories found.")
		return nil
	}

	// Convert to output type
	items := make([]repoOutput, len(repos))
	for i, r := range repos {
		items[i] = repoOutput{
			Owner:    r.Owner,
			Repo:     r.Repo,
			Path:     r.Path,
			LastSync: output.FormatTime(r.LastSyncTime(), output.TimestampISO),
		}
	}

	return output.List(p,
		[]string{"REPOSITORY", "LAST SYNC"},
		items,
		func(r repoOutput) []string {
			return []string{r.Owner + "/" + r.Repo, r.LastSync}
		},
	)
}
