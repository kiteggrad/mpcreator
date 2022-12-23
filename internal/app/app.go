package app

import (
	"os"

	"go.uber.org/zap"

	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
)

type App struct {
	mainProjectPath string
	gitlabClient    *gitlab.Client

	log *zap.SugaredLogger
}

func NewApp(mainProjectPath string, gitlabClient *gitlab.Client, log *zap.SugaredLogger) (app *App) {
	app = &App{
		mainProjectPath: mainProjectPath,
		gitlabClient:    gitlabClient,

		log: log,
	}

	return app
}

func (a *App) getSubmodulePath(submodule *git.Submodule) (submodulePath string) {
	return a.mainProjectPath + "/" + submodule.Config().Path
}

func (a *App) openMainProject() (mainProjectRepo *git.Repository, err error) {
	mainProjectRepo, err = git.PlainOpen(a.mainProjectPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to git.PlainOpen")
	}

	return mainProjectRepo, nil
}

func openDir(directoryPath string, autocreate bool, log *zap.SugaredLogger) (dir *os.File, err error) {
	dir, err = os.Open(directoryPath)
	if err != nil {
		if os.IsNotExist(err) && autocreate {
			log.Info(directoryPath, " dir is not exists, creating...")
			err = os.MkdirAll(directoryPath, os.ModePerm)
			if err != nil {
				return nil, errors.Wrap(err, "failed to os.MkdirAll")
			}
			dir, err = os.Open(directoryPath)
			if err != nil {
				return nil, errors.Wrap(err, "failed to os.Open")
			}
			log.Info(dir.Name(), " dir created")
		} else {
			return nil, errors.Wrap(err, "failed to os.Open")
		}
	} else if autocreate {
		// log.Debug(dir.Name(), " dir is exists")
	}

	{ // check directoryPath is directory
		fi, err := os.Stat(dir.Name())
		if err != nil {
			return nil, errors.Wrap(err, "failed to os.Stat")
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
		default:
			return nil, errors.Wrap(err, dir.Name()+" is not a directory")
		}
	}

	return dir, nil
}

func pointerToVar[Var any](v Var) *Var {
	return &v
}

func arrToMapKeys[K comparable](arr []K) (m map[K]struct{}) {
	m = make(map[K]struct{}, len(arr))
	for _, v := range arr {
		m[v] = struct{}{}
	}
	return m
}
