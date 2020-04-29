package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"golang.org/x/oauth2"

	"github.com/aevea/merge-master/internal/github"
	"github.com/jedib0t/go-pretty/table"
	"github.com/montanaflynn/stats"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "merge-master",
		Short: "TODO",
		Long:  "TODO",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print("There is no root command. Please check merge-master --help.")
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	oldestPRCmd := &cobra.Command{
		Use:   "pr-info",
		Short: "Gets the name and date of the longest open PR",
		RunE: func(cmd *cobra.Command, args []string) error {
			repository := cmd.Flag("repository").Value.String()
			token := cmd.Flag("token").Value.String()
			noLimit := cmd.Flag("no-limit").Value.String() == "true"

			if repository == "" {
				return errors.New("missing repository")
			}

			if token == "" {
				return errors.New("missing token")
			}

			src := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			)
			httpClient := oauth2.NewClient(context.Background(), src)

			repoName := strings.Split(repository, "/")

			githubClient := github.NewGithubClient(httpClient, repoName[0], repoName[1])

			oldestPR, err := githubClient.OldestPR()

			if err != nil {
				return err
			}

			t := table.NewWriter()

			t.AppendRow(
				table.Row{
					"Longest open PR",
					fmt.Sprintf("%.0f days", oldestPR.OpenFor.Hours()/24),
					oldestPR.URL,
				},
			)

			mergedPrs, err := githubClient.MergedPRs(noLimit)

			if err != nil {
				return err
			}

			var durations []float64

			for _, pr := range mergedPrs {
				durations = append(durations, pr.MergedAfter.Minutes())
			}

			sort.Float64s(durations)

			mean, err := stats.Mean(durations)

			if err != nil {
				return err
			}

			t.AppendRow(
				table.Row{
					fmt.Sprintf("Mean time to Merge (Last %d PRs)", len(durations)),
					fmt.Sprintf("%.0f hours", mean/60),
				},
			)

			median, err := stats.Median(durations)

			if err != nil {
				return err
			}

			t.AppendRow(
				table.Row{
					fmt.Sprintf("Median time to Merge (Last %d PRs)", len(durations)),
					fmt.Sprintf("%.0f hours", median/60),
				},
			)

			t.AppendRow(
				table.Row{
					fmt.Sprintf("Slowest time to Merge (Last %d PRs)", len(durations)),
					fmt.Sprintf("%.0f hours", durations[len(durations)-1]/60),
				},
			)

			t.AppendRow(
				table.Row{
					fmt.Sprintf("Fastest time to Merge (Last %d PRs)", len(durations)),
					fmt.Sprintf("%.2f minutes", durations[0]),
				},
			)

			fmt.Println(t.Render())

			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	oldestPRCmd.PersistentFlags().String("repository", "", "repository in the format of owner/repository")
	oldestPRCmd.PersistentFlags().String("token", "", "token for github API")
	oldestPRCmd.PersistentFlags().Bool("no-limit", false, "merge master will iterrate through all available PRs")

	rootCmd.AddCommand(oldestPRCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
