package app

import (
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (a *App) PullMainProjectSubmodules(includeGroups, excludeGroups, includProjects, excludeProjects []string) (err error) {
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
		err = a.pullSubmodule(submodule, a.log)
		if err != nil {
			return errors.Wrap(err, "failed to pullSubmodule")
		}
	}

	return nil
}

func (a *App) pullSubmodule(submodule *git.Submodule, log *zap.SugaredLogger) (err error) {
	// fmt.Println(submodule.Config().Name)
	log = log.With("submodule", submodule.Config().Name)
	log.Debug("pulling submodule...")

	submoduleRepo, err := submodule.Repository()
	if err != nil {
		return errors.Wrap(err, "failed to submodule.Repository")
	}

	submoduleCurrentBranch, err := a.getSubmoduleCurrentBranch(submodule)
	if err != nil {
		return errors.Wrap(err, "failed to submodule.Head")
	}
	// submoduleTrackingBranch gets from .git/config, not from .gitmodules - need to git sync/update or something like that?
	submoduleTrackingBranch := submodule.Config().Branch
	submoduleDefaultBranch, err := getRepoDefaultBranchName(submoduleRepo)
	if err != nil {
		return errors.Wrap(err, "failed to submodule.Head")
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
			return errors.Wrap(err, "failed to submoduleRepo.Worktree")
		}

		return nil
	}

	switch {
	case submoduleCurrentBranch == "":
		return errors.New("missing submoduleCurrentBranch")

	case submoduleTrackingBranch == "" && submoduleDefaultBranch == "":
		return errors.New("submoduleTrackingBranch and submoduleDefaultBranch are empty")

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
	submoduleRepo, err := submodule.Repository()
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
