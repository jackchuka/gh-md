package cmd

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/jackchuka/gh-md/internal/discovery"
	"github.com/spf13/cobra"
)

var reposJSON bool

var reposCmd = &cobra.Command{
	Use:   "repos",
	Short: "List all managed repositories",
	Long: `List all repositories that have been pulled with gh-md.

Shows each repository with its last sync time.

Examples:
  gh md repos              # List all repos
  gh md repos --json       # Output as JSON`,
	RunE: runRepos,
}

func init() {
	rootCmd.AddCommand(reposCmd)

	reposCmd.Flags().BoolVar(&reposJSON, "json", false, "Output as JSON")
}

func runRepos(cmd *cobra.Command, args []string) error {
	repos, err := discovery.DiscoverManagedRepos()
	if err != nil {
		return fmt.Errorf("failed to discover repositories: %w", err)
	}

	if len(repos) == 0 {
		cmd.Println("No managed repositories found.")
		return nil
	}

	if reposJSON {
		return outputReposJSON(cmd, repos)
	}

	return outputReposTable(cmd, repos)
}

func outputReposTable(cmd *cobra.Command, repos []discovery.ManagedRepo) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "REPOSITORY\tLAST SYNC")

	for _, r := range repos {
		lastSync := "-"
		if t := r.LastSyncTime(); t != nil {
			lastSync = t.Local().Format("2006-01-02 15:04:05")
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\n", r.Slug(), lastSync)
	}

	return w.Flush()
}

type repoJSONOutput struct {
	Owner    string  `json:"owner"`
	Repo     string  `json:"repo"`
	Path     string  `json:"path"`
	LastSync *string `json:"last_sync"`
}

func outputReposJSON(cmd *cobra.Command, repos []discovery.ManagedRepo) error {
	output := make([]repoJSONOutput, len(repos))
	for i, r := range repos {
		output[i] = repoJSONOutput{
			Owner: r.Owner,
			Repo:  r.Repo,
			Path:  r.Path,
		}
		if t := r.LastSyncTime(); t != nil {
			ts := t.Format("2006-01-02T15:04:05Z07:00")
			output[i].LastSync = &ts
		}
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}
