/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/kiteggrad/mpcreator/internal/app"
	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

// mpfillCmd represents the mpfill command
var mpfillCmd = &cobra.Command{
	Use:   "mpfill",
	Short: "Заполняет главный проект",
	Long: `Заполняет главный проект:
создаёт (mkdir, git init) главный репозиторий (если его нет) по указанному пути,
добавляет туда все репозитории из гитлаба как подмодули. 
Если репозиторий уже есть - просто добавляет его в список подмодулей (не трогаает изменения)`,
	Example: `mpcreator mpfill -p /home/derbenev/go/src/project -u https://gitlab.ru -t yourToken`,

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
		includeLanguages, err := cmd.Flags().GetStringSlice("inlang")
		if err != nil {
			return errors.Wrap(err, "failed to get inlang flag")
		}
		excludeLanguages, err := cmd.Flags().GetStringSlice("exlang")
		if err != nil {
			return errors.Wrap(err, "failed to get exlang flag")
		}
		log.Printf("includeGroups: %#+v\n", includeGroups)
		log.Printf("excludeGroups: %#+v\n", excludeGroups)

		gitlabClient, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabURL))
		if err != nil {
			return errors.Wrap(err, "failed to gitlab.NewClient")
		}

		app := app.NewApp(mainProjectPath, gitlabClient, zap.S())
		err = app.FillMainProject(
			includeGroups, excludeGroups,
			includeProjects, excludeProjects,
			includeLanguages, excludeLanguages,
		)
		if err != nil {
			return errors.Wrap(err, "failed to app.FillMainProject")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(mpfillCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mpfillCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mpfillCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	mpfillCmd.Flags().StringP("mppath", "p", "", "path to main project e.g. /home/derbenev/go/src/rnis")
	mpfillCmd.MarkFlagRequired("mppath")
	mpfillCmd.MarkFlagDirname("mppath")

	mpfillCmd.Flags().StringP("url", "u", "", "gitlab url e.g. https://gitlab.ru")
	mpfillCmd.MarkFlagRequired("url")

	mpfillCmd.Flags().StringP("token", "t", "", "gitlab api token")
	mpfillCmd.MarkFlagRequired("token")

	mpfillCmd.Flags().StringSlice("ingroups", nil, `included groups e.g. "etp" or "etp,etp/parser"`)
	mpfillCmd.Flags().StringSlice("exgroups", nil, `excluded groups e.g. "etp" or "etp,etp/parser"`)
	mpfillCmd.Flags().StringSlice("inprojects", nil, `included projects e.g. "events-geo" or "events-overspeed,events-geo"`)
	mpfillCmd.Flags().StringSlice("exprojects", nil, `excluded projects e.g. "events-geo" or "events-overspeed,events-geo"`)
	mpfillCmd.Flags().StringSlice("inlang", nil, `included languages e.g. "Go" or "Go,CSS"`)
	mpfillCmd.Flags().StringSlice("exlang", nil, `excluded languages e.g. "Go" or "Go,CSS"`)
}
