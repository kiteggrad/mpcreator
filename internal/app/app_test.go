package app

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/goleak"
	"go.uber.org/zap"
)

const (
	EnvGitlabURL   = "TEST_GITLAB_URL"
	EnvGitlabToken = "TEST_GITLAB_TOKEN"
)

func TestApp(t *testing.T) {
	// t.Skip()
	defer goleak.VerifyNone(t)

	suite.Run(t, new(AppTestSuite))
}

type AppTestSuite struct {
	suite.Suite

	mainProjectPath string
	gitlabToken     string
	gitlabURL       string

	gitlabClient *gitlab.Client
	log          *zap.SugaredLogger

	app *App
}

// you can add your test env vars to go.testEnvVars (.vscode/settings.json)
func bindEnv() {
	err := viper.BindEnv(EnvGitlabURL)
	if err != nil {
		panic(err)
	}
	err = viper.BindEnv(EnvGitlabToken)
	if err != nil {
		panic(err)
	}
}

func (s *AppTestSuite) SetupSuite() {
	bindEnv()

	s.gitlabToken = viper.GetString(EnvGitlabToken)
	s.gitlabURL = viper.GetString(EnvGitlabURL)

	if s.gitlabURL == "" {
		s.T().Skip("empty gitlabURL")
	}

	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.Development = false

	log, err := zapCfg.Build()
	if !s.NoError(err) {
		s.T().FailNow()
	}
	s.log = log.Sugar()
	zap.ReplaceGlobals(log)

	pwd, err := os.Getwd()
	if !s.NoError(err) {
		s.T().FailNow()
	}
	s.mainProjectPath = pwd + "/test"
	mainProjectDir, err := openDir(s.mainProjectPath, true, zap.S())
	s.T().Cleanup(func() {
		_ = mainProjectDir
		os.RemoveAll(mainProjectDir.Name())
	})

	s.gitlabClient, err = gitlab.NewClient(
		s.gitlabToken, gitlab.WithBaseURL(s.gitlabURL),
		gitlab.WithHTTPClient(http.DefaultClient),
	)
	s.NoError(err)

	s.app = NewApp(s.mainProjectPath, s.gitlabClient, log.Sugar())
}
func (s *AppTestSuite) TearDownSuite() {
	http.DefaultClient.CloseIdleConnections()
}

func (s *AppTestSuite) SetupTest() {
}
func (s *AppTestSuite) TearDownTest() {
}

func (s *AppTestSuite) Test_FillMainProject() {
	s.T().Skip()

	err := s.app.FillMainProject(
		[]string{"rupor"}, []string{}, // groups in / ex
		[]string{"rupor-search-microservice"}, []string{}, // projects in / ex
		[]string{"Go"}, []string{}, // languages in / ex
	)
	s.NoError(err)
}

func (s *AppTestSuite) Test_PullMainProjectSubmodules() {
	s.T().Skip()

	err := s.app.PullMainProjectSubmodules(nil, nil, nil, nil)
	s.NoError(err)
}

func (s *AppTestSuite) Test_AddSubmodule() {
	s.T().Skip()

	mainRepo, err := s.app.initMainProject()
	s.NoError(err)

	submodule, err := addSubmoduleToRepo(
		mainRepo,
		"rupor/rupor-search-microservice",
		"git@gitlab.cyrm.ru:rupor/rupor-search-microservice.git",
		s.log,
	)
	s.NoError(err)

	_ = submodule
}

func (s *AppTestSuite) Test_InitMainProject() {
	s.T().Skip()

	mainRepo, err := s.app.initMainProject()
	s.NoError(err)

	_ = mainRepo
}

func (s *AppTestSuite) Test_IterateGroupsProjects() {
	s.T().Skip()

	err := s.app.iterateGroupsProjects(
		func(group *gitlab.Group) (err error) {
			fmt.Println("group", group.FullPath, group.WebURL)
			return nil
		},
		func(project *gitlab.Project) (err error) {
			languages, _, err := s.gitlabClient.Projects.GetProjectLanguages(project.ID)
			if err != nil {
				return err
			}
			fmt.Println("project", project.PathWithNamespace, project.SSHURLToRepo, languages)
			return nil
		},
		[]string{"rupor"}, []string{}, // groups in / ex
		[]string{}, []string{}, // projects in / ex
		[]string{"Go"}, []string{}, // languages in / ex
	)
	s.NoError(err)
}

func (s *AppTestSuite) Test_FindDirectory() {
	s.T().Skip()

	dir, err := openDir(s.mainProjectPath, true, zap.S())
	s.NoError(err)
	_ = dir
}
