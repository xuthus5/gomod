package gomod

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/google/go-github/v64/github"
	"github.com/olekukonko/tablewriter"
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

	whiteList = []string{"golang.org"}
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

func GetModFile(mf string) (*modfile.File, error) {
	modFileData, err := os.ReadFile(mf)
	if err != nil {
		logrus.Errorf("read go.mod file failed: %v", err)
		return nil, err
	}
	modFile, err := modfile.Parse("go.mod", modFileData, nil)
	if err != nil {
		logrus.Errorf("parse go.mod file failed: %v", err)
		return nil, err
	}
	return modFile, nil
}

func ModUpgrade(upgradeIndirect bool) {
	modFile, err := GetModFile("go.mod")
	if err != nil {
		logrus.Errorf("get go.mod file failed: %v", err)
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

type Module struct {
	Path       string       // module path
	Query      string       // version query corresponding to this version
	Version    string       // module version
	Versions   []string     // available module versions
	Replace    *Module      // replaced by this module
	Time       *time.Time   // time version was created
	Update     *Module      // available update (with -u)
	Main       bool         // is this the main module?
	Indirect   bool         // module is only indirectly needed by main module
	Dir        string       // directory holding local copy of files, if any
	GoMod      string       // path to go.mod file describing module, if any
	GoVersion  string       // go version used in module
	Retracted  []string     // retraction information, if any (with -retracted or -u)
	Deprecated string       // deprecation message, if any (with -u)
	Error      *ModuleError // error loading module
	Origin     any          // provenance of module
	Reuse      bool         // reuse of old module info is safe
}

type ModuleError struct {
	Err string // the error itself
}

func Analyzed() {
	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spin.Suffix = " Analyzing for dependencies..."
	spin.Start()

	modules, err := analyzed("go", "list", "-m", "-json", "-mod=readonly", "all")
	if err != nil {
		spin.Stop()
		logrus.Errorf("analyzed project dependencies failed: %v", err)
		return
	}

	spin.Stop()

	var tableRows [][]string
	for _, mod := range modules {
		tableRows = append(tableRows, []string{
			mod.Path,
			getRelation(mod),
			mod.Version,
			getGoVersion(mod),
			getToolChains(mod),
		})
	}

	if len(tableRows) == 0 {
		awesome := color.New(color.FgHiGreen, color.Bold).Sprint("✔ Empty!")
		fmt.Printf(" %s All of your dependencies are empty.\n", awesome)
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Package", "Relation", "Version", "GoVersion", "ToolChains"})
		table.EnableBorder(true)
		table.AppendBulk(tableRows)
		table.Render()
	}
}

func getGoVersion(m *Module) string {
	if m.GoVersion != "" {
		return m.GoVersion
	}
	var mf string
	if m.GoMod != "" {
		mf = m.GoMod
	} else if m.Dir != "" {
		mf = filepath.Join(m.Dir, "go.mod")
	}
	if mf == "" {
		return ""
	}
	modFile, err := GetModFile(mf)
	if err != nil {
		return ""
	}
	if modFile.Go != nil {
		return modFile.Go.Version
	}
	return ""
}

func getRelation(m *Module) string {
	if m.Main {
		return "main"
	}
	if m.Indirect {
		return "indirect"
	}
	return "direct"
}

func getToolChains(m *Module) string {
	var mf string
	if m.GoMod != "" {
		mf = m.GoMod
	} else if m.Dir != "" {
		mf = filepath.Join(m.Dir, "go.mod")
	}
	if mf == "" {
		return ""
	}
	file, err := GetModFile(mf)
	if err != nil {
		return ""
	}
	if file.Toolchain == nil {
		return ""
	}
	return file.Toolchain.Name
}

func analyzed(cmd string, args ...string) ([]*Module, error) {
	execBuf, err := execute(cmd, args...)
	if err != nil {
		logrus.Errorf("go list failed: %s, return: %v", execBuf, err)
		return nil, err
	}
	var modules []*Module
	var buf bytes.Buffer
	var depth int32
	for _, ch := range execBuf {
		switch ch {
		case '{':
			depth++
			buf.WriteByte(ch)
		case '}':
			depth--
			if depth == 0 {
				buf.WriteByte(ch)
				var m = new(Module)
				if err := json.Unmarshal(buf.Bytes(), m); err != nil {
					return nil, err
				}
				buf.Reset()
				modules = append(modules, m)
			} else {
				buf.WriteByte(ch)
			}
		default:
			buf.WriteByte(ch)
		}
	}
	return modules, nil
}

func execute(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	return cmd.CombinedOutput()
}

func UpdateList() {
	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spin.Suffix = " Checking for updates..."
	spin.Start()

	modules, err := analyzed("go", "list", "-m", "-u", "-json", "-mod=readonly", "all")
	if err != nil {
		spin.Stop()
		logrus.Errorf("analyzed project dependencies failed: %v", err)
		return
	}

	spin.Stop()

	var tableRows [][]string
	for _, mod := range modules {
		if mod.Update == nil {
			continue
		}
		tableRows = append(tableRows, []string{
			mod.Path,
			getRelation(mod),
			mod.Version,
			mod.Update.Version,
		})
	}

	if len(tableRows) == 0 {
		awesome := color.New(color.FgHiGreen, color.Bold).Sprint("✔ Empty!")
		fmt.Printf(" %s All of your dependencies are empty.\n", awesome)
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Package", "Relation", "Current", "Latest"})
		table.EnableBorder(true)
		table.AppendBulk(tableRows)
		table.Render()
	}
}
