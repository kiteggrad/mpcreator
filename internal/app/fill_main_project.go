package app

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (a *App) FillMainProject(
	includeGroups, excludeGroups,
	includeProjects, excludeProjects,
	includeLanguages, excludeLanguages []string,
) (err error) {
	mainProjectRepo, err := a.initMainProject()
	if err != nil {
		return errors.Wrap(err, "failed to initMainProject")
	}

	err = a.iterateGroups(func(group *gitlab.Group) (err error) {
		log := a.log.With("group", group.FullPath)
		log.Debug("filling group ...")

		g := &errgroup.Group{}

		err = a.iterateGroupProjects(group, func(project *gitlab.Project) (err error) {
			g.Go(func() (err error) {
				log := log.With("project", project.PathWithNamespace)
				log.Debug("filling project ...")
				defer log.Debug("filling project done")

				projectPath := strings.ReplaceAll(project.PathWithNamespace, " / ", "/")
				_, err = addSubmoduleToRepo(mainProjectRepo, projectPath, project.SSHURLToRepo, log)
				if err != nil {
					log.With(zap.Error(err)).Error("failed to addSubmoduleToRepo")
				}
				return nil
			})

			return nil
		},
			includeProjects, excludeProjects,
			includeLanguages, excludeLanguages,
		)
		if err != nil {
			return errors.Wrap(err, "failed to iterateGroupProjects")
		}

		err = g.Wait()
		if err != nil {
			return errors.Wrap(err, "failed to g.Wait")
		}

		return nil
	}, includeGroups, excludeGroups)
	if err != nil {
		return errors.Wrap(err, "failed to iterateGroups")
	}

	return nil
}

func (a *App) initMainProject() (mainProjectRepo *git.Repository, err error) {
	dir, err := openDir(a.mainProjectPath, true, a.log)
	if err != nil {
		panic(errors.Wrap(err, "failed to openDir"))
	}

	mainProjectRepo, err = git.PlainOpen(dir.Name())
	switch err {
	case git.ErrRepositoryNotExists:
		a.log.Info(a.mainProjectPath, " repo is not exists, creating ...")

		mainProjectRepo, err = git.PlainInit(a.mainProjectPath, false)
		if err != nil {
			return nil, errors.Wrap(err, "failed to git.PlainInit")
		}

		a.log.Info(a.mainProjectPath, " repo created")

	case nil:
		a.log.Debug(a.mainProjectPath, " repo is exists")

	default:
		return nil, errors.Wrap(err, "failed to git.PlainOpen")
	}

	return mainProjectRepo, nil
}

func (a *App) iterateGroupsProjects(
	groupCallback func(group *gitlab.Group) (err error),
	projectCallback func(project *gitlab.Project) (err error),
	includeGroups, excludeGroups,
	includProjects, excludeProjects,
	includeLanguages, excludeLanguages []string,
) (err error) {
	err = a.iterateGroups(func(group *gitlab.Group) (err error) {
		if groupCallback != nil {
			err = groupCallback(group)
			if err != nil {
				return errors.Wrap(err, "failed to groupCallback")
			}
		}
		if projectCallback != nil {
			err = a.iterateGroupProjects(
				group, projectCallback,
				includProjects, excludeProjects,
				includeLanguages, excludeLanguages,
			)
			if err != nil {
				return errors.Wrap(err, "failed to iterateGroupProjects")
			}
		}
		return nil
	}, includeGroups, excludeGroups)
	if err != nil {
		return errors.Wrap(err, "failed to iterateGroups")
	}
	return nil
}

func (a *App) iterateGroups(
	groupCallback func(group *gitlab.Group) (err error),
	include, exclude []string,
) (err error) {
	var search *string
	if len(include) == 1 {
		search = pointerToVar(include[0])
	}

	for page, perPage := 1, 100; ; page++ {
		groups, _, err := a.gitlabClient.Groups.ListGroups(&gitlab.ListGroupsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
			AllAvailable: pointerToVar(false),
			TopLevelOnly: pointerToVar(false),
			Search:       search,
		})
		if err != nil {
			return errors.Wrap(err, "failed to client.Groups.ListGroups")
		}

		for _, group := range groups {
			if !groupIncludeExcludePass(include, exclude, group) {
				continue
			}

			err = groupCallback(group)
			if err != nil {
				return errors.Wrap(err, "failed to groupCallback")
			}
		}

		if len(groups) != perPage {
			break
		}
	}

	return nil
}

func groupIncludeExcludePass(includeGroups, excludeGroups []string, group *gitlab.Group) (pass bool) {
	formattedGroupName := "/" + strings.ReplaceAll(group.FullPath, " / ", "/") + "/"
	include, exclude := true, false

	if len(includeGroups) != 0 {
		include = false
		for _, includeGroup := range includeGroups {
			if strings.Contains(formattedGroupName, "/"+includeGroup+"/") {
				include = true
				break
			}
		}
	}
	if len(excludeGroups) != 0 {
		for _, excludeGroup := range excludeGroups {
			if strings.Contains(formattedGroupName, "/"+excludeGroup+"/") {
				exclude = true
				break
			}
		}
	}

	return include && (!exclude)
}

func (a *App) projectIncludeExcludePass(
	includeProjects, excludeProjects,
	includeLanguages, excludeLanguages []string,
	project *gitlab.Project,
) (pass bool, err error) {
	// projects
	{
		if len(includeProjects) != 0 { // include
			include := false
			for _, includeProject := range includeProjects {
				if project.Path == includeProject {
					include = true
					break
				}
			}
			if !include {
				return false, nil
			}
		}

		if len(excludeProjects) != 0 { // exclude
			for _, excludeProject := range excludeProjects {
				if project.Path == excludeProject {
					return false, nil
				}
			}
		}
	}

	// languages
	if len(includeLanguages) != 0 || len(excludeLanguages) != 0 {
		languages, _, err := a.gitlabClient.Projects.GetProjectLanguages(project.ID)
		if err != nil {
			return false, errors.Wrapf(err, "failed to GetProjectLanguages for project.ID %d", project.ID)
		}

		if len(includeLanguages) != 0 { // include
			include := false
			for _, includeLanguage := range includeLanguages {
				if _, ok := (*languages)[includeLanguage]; ok {
					include = true
					break
				}
			}
			if !include {
				return false, nil
			}
		}

		if len(excludeLanguages) != 0 { // exclude
			for _, excludeLanguage := range excludeLanguages {
				if _, ok := (*languages)[excludeLanguage]; ok {
					return false, nil
				}
			}
		}
	}

	return true, nil
}

func (a *App) iterateGroupProjects(
	group *gitlab.Group,
	projectCallback func(project *gitlab.Project) (err error),
	includeProjects, excludeProjects,
	includeLanguages, excludeLanguages []string,
) (err error) {
	for page, perPage := 1, 100; ; page++ {
		groupProjects, _, err := a.gitlabClient.Groups.ListGroupProjects(group.ID, &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
			Archived:         pointerToVar(false),
			IncludeSubGroups: pointerToVar(false),
		})
		if err != nil {
			return errors.Wrap(err, "failed to client.Groups.ListGroupProjects")
		}

		for _, groupProject := range groupProjects {
			pass, err := a.projectIncludeExcludePass(
				includeProjects, excludeProjects,
				includeLanguages, excludeLanguages,
				groupProject,
			)
			if err != nil {
				return errors.Wrap(err, "failed to projectIncludeExcludePass")
			}

			if !pass {
				continue
			}
			err = projectCallback(groupProject)
			if err != nil {
				return errors.Wrap(err, "failed to projectCallback")
			}
		}

		if len(groupProjects) != perPage {
			break
		}
	}

	return nil
}

func addSubmoduleToRepo(
	repo *git.Repository,
	submodulePath,
	submoduleURL string,
	log *zap.SugaredLogger,
) (submodule *git.Submodule, err error) {
	wt, err := repo.Worktree()
	if err != nil {
		return nil, errors.Wrap(err, "failed to repo.Worktree")
	}
	log = log.With(
		"submodulePath", submodulePath,
		"submoduleURL", submoduleURL,
	)

	submodule, err = wt.Submodule(submodulePath)
	if err != nil && errors.Is(err, git.ErrSubmoduleNotFound) {
		log.Info("submodule not exists, creating ...")

		args := []string{"submodule", "add", submoduleURL}
		if submodulePath != "" {
			args = append(args, submodulePath)
		}
		cmd := exec.Command("git", args...)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "LC_ALL", "C")
		cmd.Dir = wt.Filesystem.Root()
		// var out bytes.Buffer
		var stderr bytes.Buffer
		// cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return nil, errors.Wrap(err, "failed to exec.Command(git submodule add) "+stderr.String())
		}

		log.Info("submodule created")

		submodule, err = wt.Submodule(submodulePath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to wt.Submodule after submodule add")
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to wt.Submodule")
	} else {
		log.Debug("submodule exists")
	}

	err = submodule.Init()
	if err != nil && !errors.Is(err, git.ErrSubmoduleAlreadyInitialized) {
		return nil, errors.Wrap(err, "failed to submodule.Init")
	}

	return submodule, nil
}
