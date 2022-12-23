/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/kiteggrad/mpcreator/internal/app"
	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Подтягивает последние изменения по репозиториям",
	Long: `Подтягивает последние изменения по репозиториям:
пробегается по всем репозиториям и делает git pull,
применится только для репозиториев у которых текущая ветка == основной ветке
иначе репозиторий будет пропущен.`,
	Example: `mpcreator pull -p /home/derbenev/go/src/project -u https://gitlab.ru -t yourToken`,

	RunE: func(cmd *cobra.Command, args []string) error {
		mainProjectPath := cmd.Flags().Lookup("mppath").Value.String()
		gitlabURL := cmd.Flags().Lookup("url").Value.String()
		gitlabToken := cmd.Flags().Lookup("token").Value.String()
		includeGroups, err := cmd.Flags().GetStringSlice("ingroups")
		if err != nil {
			return errors.Wrap(err, "failed to get ingroup flag")
		}
		excludeGroups, err := cmd.Flags().GetStringSlice("exgroups")
		if err != nil {
			return errors.Wrap(err, "failed to get exgroup flag")
		}
		includeProjects, err := cmd.Flags().GetStringSlice("inprojects")
		if err != nil {
			return errors.Wrap(err, "failed to get inproject flag")
		}
		excludeProjects, err := cmd.Flags().GetStringSlice("exprojects")
		if err != nil {
			return errors.Wrap(err, "failed to get exproject flag")
		}
		// includeLanguages, err := cmd.Flags().GetStringSlice("inlang")
		// if err != nil {
		// 	return errors.Wrap(err, "failed to get inlang flag")
		// }
		// excludeLanguages, err := cmd.Flags().GetStringSlice("exlang")
		// if err != nil {
		// 	return errors.Wrap(err, "failed to get exlang flag")
		// }

		gitlabClient, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabURL))
		if err != nil {
			return errors.Wrap(err, "failed to gitlab.NewClient")
		}

		app := app.NewApp(mainProjectPath, gitlabClient, zap.S())
		err = app.PullMainProjectSubmodules(
			includeGroups, excludeGroups,
			includeProjects, excludeProjects,
			// includeLanguages, excludeLanguages,
		)
		if err != nil {
			return errors.Wrap(err, "failed to app.PullMainProjectSubmodules")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pullCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pullCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	pullCmd.Flags().StringP("mppath", "p", "", "path to main project e.g. /home/derbenev/go/src/rnis")
	pullCmd.MarkFlagRequired("mppath")
	pullCmd.MarkFlagDirname("mppath")

	pullCmd.Flags().StringP("url", "u", "", "gitlab url e.g. https://gitlab.ru")
	pullCmd.MarkFlagRequired("url")

	pullCmd.Flags().StringP("token", "t", "", "gitlab api token")
	pullCmd.MarkFlagRequired("token")

	pullCmd.Flags().StringSlice("ingroups", nil, `included groups e.g. "etp" or "etp,etp/parser"`)
	pullCmd.Flags().StringSlice("exgroups", nil, `excluded groups e.g. "etp" or "etp,etp/parser"`)
	pullCmd.Flags().StringSlice("inprojects", nil, `included projects e.g. "events-geo" or "events-overspeed,events-geo"`)
	pullCmd.Flags().StringSlice("exprojects", nil, `excluded projects e.g. "events-geo" or "events-overspeed,events-geo"`)
	// pullCmd.Flags().StringSlice("inlang", nil, `included languages e.g. "Go" or "Go,CSS"`)
	// pullCmd.Flags().StringSlice("exlang", nil, `excluded languages e.g. "Go" or "Go,CSS"`)
}
