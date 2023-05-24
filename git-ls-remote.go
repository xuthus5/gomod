package gomod

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	lsRemote _lsRemote

	tagRegexp  = regexp.MustCompile(`refs/tags/(.*)`)
	headRegexp = regexp.MustCompile(`refs/heads/(master|main)`)
)

type _lsRemote struct {
	url    string
	output string
}

func (lr *_lsRemote) setUrl(url string) *_lsRemote {
	url = strings.ReplaceAll(url, "https://", "")
	url = strings.ReplaceAll(url, "http://", "")
	if v := strings.Split(url, "/"); len(v) > 3 {
		url = strings.Join(v[:3], "/")
	}
	if !strings.HasPrefix(url, "https") {
		url = "https://" + url
	}
	lr.url = url
	return lr
}

func (lr *_lsRemote) command() error {
	cmd := exec.Command("git", "ls-remote", "--heads", "--tags", "--sort=v:refname", lr.url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.WithField("cmd", cmd.String()).Errorf("run command failed: %v", err)
		return err
	}
	lr.output = string(output)
	return nil
}

func (lr *_lsRemote) tagOrCommitID() (string, error) {
	if err := lr.command(); err != nil {
		return "", err
	}
	lr.output = strings.Trim(lr.output, "\n ")
	commits := strings.Split(lr.output, "\n")
	if len(commits) == 0 {
		return "", fmt.Errorf("get commit log empty")
	}
	commit := commits[len(commits)-1]

	if strings.Contains(commit, "refs/heads/") {
		matches := headRegexp.FindStringSubmatch(commit)
		if len(matches) != 2 {
			return "", fmt.Errorf("%s no match rule: %s", commit, headRegexp.String())
		}
		return strings.Trim(strings.ReplaceAll(commit, matches[0], ""), " \n\t\r"), nil
	}

	if strings.Contains(commit, "refs/tags/") {
		matches := tagRegexp.FindStringSubmatch(commit)
		if len(matches) != 2 {
			return "", fmt.Errorf("%s no match rule: %s", commit, tagRegexp.String())
		}
		return matches[1], nil
	}

	return "", fmt.Errorf("%s no match heads and tags rule", commit)
}
