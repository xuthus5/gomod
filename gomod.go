package gomod

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-github/v64/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/mod/modfile"
	"golang.org/x/net/html"
)

const (
	defaultVersion = "latest"
)

var (
	httpClient = &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyFromEnvironment},
		Timeout:   time.Second * 5,
	}
)

func goGet(u, v string) error {
	versionUrl := u + "@" + v
	cmd := exec.Command("go", "get", "-u", versionUrl)
	if _, err := cmd.CombinedOutput(); err != nil {
		return err
	}
	return nil
}

type upgrade interface {
	upgrade() error
}

type githubRepo struct {
	client *github.Client
	url    string
	owner  string
	repo   string
	semV   string
}

func newGithubRepo(url string) *githubRepo {
	return &githubRepo{
		client: github.NewClient(&http.Client{
			Transport: &http.Transport{Proxy: http.ProxyFromEnvironment},
			Timeout:   time.Second * 3,
		}),
		url: url,
	}
}

func (g *githubRepo) parse() error {
	ss := strings.Split(g.url, "/")
	if len(ss) < 3 {
		return fmt.Errorf("invalid github url: %s", g.url)
	}
	g.owner = ss[1]
	g.repo = ss[2]
	if len(ss) == 4 {
		g.semV = ss[3]
	}

	return nil
}

func (g *githubRepo) getVersion() string {
	ctx := context.Background()
	release, _, err := g.client.Repositories.GetLatestRelease(ctx, g.owner, g.repo)
	if err == nil {
		return *release.TagName
	}
	// get latest commit id
	commits, _, err := g.client.Repositories.ListCommits(ctx, g.owner, g.repo, &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 1,
		},
	})
	if err != nil {
		return defaultVersion
	}
	if len(commits) == 0 {
		return defaultVersion
	}
	return (*commits[0].SHA)[:8]
}

func (g *githubRepo) upgrade() error {
	if err := g.parse(); err != nil {
		return err
	}
	v := g.getVersion()
	if err := goGet(g.url, v); err != nil {
		return err
	}
	logrus.WithField("url", g.url+"@"+v).Infof("upgrade success")
	return nil
}

type repo struct {
	dependency string
	originRepo string
}

func newRepo(dep string) *repo {
	return &repo{
		dependency: dep,
	}
}

func (r *repo) extraGoImportMetadata(reader io.Reader) (string, error) {
	tokenizer := html.NewTokenizer(reader)
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			err := tokenizer.Err()
			if errors.Is(err, io.EOF) {
				return "", errors.New("go-import meta attr not found")
			}
			return "", errors.New("read dependency metadata failed: " + err.Error())
		case html.StartTagToken, html.SelfClosingTagToken:
			t := tokenizer.Token()
			if t.Data != "meta" {
				continue
			}
			for i, attribute := range t.Attr {
				if attribute.Key == "name" && attribute.Val == "go-import" && i+1 < len(t.Attr) {
					return t.Attr[i+1].Val, nil
				}
			}
		default:
			continue
		}
	}
}

func (r *repo) upgrade() error {
	u := "https://" + r.dependency + "?go-get=1"
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	metadata, err := r.extraGoImportMetadata(response.Body)
	if err != nil {
		return err
	}
	ss := strings.Split(metadata, " ")
	if len(ss) != 3 {
		return errors.New("go-import metadata invalid: " + metadata)
	}
	r.originRepo = ss[2]

	v, err := lsRemote.setUrl(r.originRepo).tagOrCommitID()
	if err != nil {
		return err
	}
	if err := goGet(r.dependency, v); err != nil {
		return err
	}
	logrus.WithField("url", r.dependency+"@"+v).Infof("upgrade success")
	return nil
}

func ModUpgrade(upgradeIndirect bool) {
	modFileData, err := os.ReadFile("go.mod")
	if err != nil {
		logrus.Errorf("read go.mod file failed: %v", err)
		return
	}
	modFile, err := modfile.Parse("go.mod", modFileData, nil)
	if err != nil {
		logrus.Errorf("parse go.mod file failed: %v", err)
		return
	}
	for _, mod := range modFile.Require {
		if mod.Indirect && !upgradeIndirect {
			continue
		}
		var repo upgrade
		if strings.Contains(mod.Mod.Path, "github.com") {
			repo = newGithubRepo(mod.Mod.Path)
		} else {
			repo = newRepo(mod.Mod.Path)
		}
		if err := repo.upgrade(); err != nil {
			logrus.WithField("url", mod.Mod.Path).Errorf("upgrade failed: %v. (starting fallback)", err)
			fallback(mod.Mod.Path)
			continue
		}
	}

	tidy()
}

func tidy() {
	cmd := exec.Command("go", "mod", "tidy")
	_ = cmd.Run()
}

func fallback(repo string) {
	if err := goGet(repo, defaultVersion); err != nil {
		logrus.WithField("url", repo+"@"+defaultVersion).Errorf("upgrade failed: %v", err)
		return
	}
	logrus.WithField("url", repo+"@"+defaultVersion).Infof("upgrade success")
}
