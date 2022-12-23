package app

import (
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

func (a *App) PullMainProjectSubmodules(includeGroups, excludeGroups, includeProjects, excludeProjects []string) (err error) {
	a.log.With(
		"includeGroups", includeGroups, "excludeGroups", excludeGroups,
		"includeProjects", includeProjects, "excludeProjects", excludeProjects,
	).Info("PullMainProjectSubmodules")

	mainProjectRepo, err := a.openMainProject()
	if err != nil {
		return errors.Wrap(err, "failed to openMainProject")
	}
	wt, err := mainProjectRepo.Worktree()
	if err != nil {
		return errors.Wrap(err, "failed to mainProjectRepo.Worktree")
	}
	submodules, err := wt.Submodules()
	if err != nil {
		return errors.Wrap(err, "failed to wt.Submodules")
	}

	for _, submodule := range submodules {
		submoduleFullName := submodule.Config().Name
		submoduleGroupName := strings.Split(submoduleFullName, "/")[0]
		submoduleProjectName := strings.Split(submoduleFullName, "/")[1]

		if len(includeGroups) != 0 { // TODO move to func
			if !slices.Contains(includeGroups, submoduleGroupName) {
				continue
			}
		}
		if len(excludeGroups) != 0 {
			if slices.Contains(excludeGroups, submoduleGroupName) {
				continue
			}
		}

		if len(includeProjects) != 0 {
			if !slices.Contains(includeProjects, submoduleProjectName) {
				continue
			}
		}
		if len(excludeProjects) != 0 {
			if slices.Contains(excludeProjects, submoduleProjectName) {
				continue
			}
		}

		err = a.pullSubmodule(submodule, a.log)
		if err != nil {
			a.log.With(
				"submodule", submodule.Config().Path,
				"error", err.Error(),
			).Error("failed to pullSubmodule")

			continue
		}
	}

	return nil
}

func (a *App) pullSubmodule(submodule *git.Submodule, log *zap.SugaredLogger) (err error) {
	// fmt.Println(submodule.Config().Name)
	log = log.With("submodule", submodule.Config().Name)
	log.Debug("pulling submodule...")

	// not submodule.Repository() because it randomly throws error
	submoduleRepo, err := git.PlainOpen(a.getSubmodulePath(submodule))
	if err != nil {
		return errors.Wrap(err, "failed to submodule.Repository")
	}

	submoduleCurrentBranch, err := a.getSubmoduleCurrentBranch(submodule)
	if err != nil {
		return errors.Wrap(err, "failed to getSubmoduleCurrentBranch")
	}
	// submoduleTrackingBranch gets from .git/config, not from .gitmodules - need to git sync/update or something like that?
	submoduleTrackingBranch := submodule.Config().Branch
	submoduleDefaultBranch, err := getRepoDefaultBranchName(submoduleRepo)
	if err != nil {
		return errors.Wrap(err, "failed to getRepoDefaultBranchName")
	}

	log = log.With(
		"submoduleCurrentBranch", submoduleCurrentBranch,
		"submoduleTrackingBranch", submoduleTrackingBranch,
		"submoduleDefaultBranch", submoduleDefaultBranch,
	)

	pull := func() error {
		submoduleWorktree, err := submoduleRepo.Worktree()
		if err != nil {
			return errors.Wrap(err, "failed to submoduleRepo.Worktree")
		}

		err = submoduleWorktree.Pull(&git.PullOptions{RemoteName: "origin"})
		switch err {
		case git.NoErrAlreadyUpToDate:
		case nil:
			log.Info("pulled new changes")
		default:
			return errors.Wrap(err, "failed to submoduleWorktree.Pull")
		}

		return nil
	}

	switch {
	case submoduleCurrentBranch == "":
		return errors.New("missing submoduleCurrentBranch")

	case submoduleTrackingBranch == "" && submoduleDefaultBranch == "":
		return errors.New("submoduleTrackingBranch and submoduleDefaultBranch are empty" +
			". You can try to fix it by 'git symbolic-ref refs/remotes/origin/HEAD refs/remotes/origin/YOUR_DEFAULT_BRANCH'",
		)

	case submoduleTrackingBranch != "": // pull from submoduleTrackingBranch
		if submoduleTrackingBranch != submoduleCurrentBranch {
			log.Warn("submoduleTrackingBranch != submoduleCurrentBranch, skipping")
			return nil
		}

		err = pull()
		if err != nil {
			return errors.Wrap(err, "failed to pull submoduleTrackingBranch")
		}

	case submoduleDefaultBranch != "": // pull from submoduleDefaultBranch
		if submoduleDefaultBranch != submoduleCurrentBranch {
			log.Warn("submoduleDefaultBranch != submoduleCurrentBranch, skipping")
			return nil
		}

		err = pull()
		if err != nil {
			return errors.Wrap(err, "failed to pull submoduleDefaultBranch")
		}

	default:
		log.Panic("unexpected case")
	}

	return nil
}

func (a *App) getSubmoduleCurrentBranch(submodule *git.Submodule) (submoduleCurrentBranch string, err error) {
	// not submodule.Repository() because it randomly throws error
	submoduleRepo, err := git.PlainOpen(a.getSubmodulePath(submodule))
	if err != nil {
		return "", errors.Wrap(err, "failed to submodule.Repository")
	}
	submoduleRepoHead, err := submoduleRepo.Head()
	if err != nil {
		return "", errors.Wrap(err, "failed to submodule.Head")
	}

	return submoduleRepoHead.Name().Short(), nil

	// cmd := exec.Command("git", "symbolic-ref", "-q", "HEAD")
	// cmd.Env = append(cmd.Env, "LC_ALL", "C")
	// cmd.Dir = a.getSubmodulePath(submodule)
	// var out bytes.Buffer
	// var stderr bytes.Buffer
	// cmd.Stdout = &out
	// cmd.Stderr = &stderr
	// err = cmd.Run()
	// if err != nil {
	// 	return "", errors.Wrap(err, "failed to exec.Command(git symbolic-ref -q HEAD) "+stderr.String())
	// }

	// return plumbing.ReferenceName(out.String()).Short(), nil
}

func getRepoDefaultBranchName(repo *git.Repository) (name string, err error) {
	t, err := repo.References()
	if err != nil {
		return "", errors.Wrap(err, "failed to repo.References")
	}

	t.ForEach(func(r *plumbing.Reference) error {
		if r.Type() == plumbing.SymbolicReference && r.Name().Short() == "origin/HEAD" {
			name = strings.TrimPrefix(r.Target().Short(), "origin/")
			return nil
		}
		return nil
	})

	return name, nil
}
