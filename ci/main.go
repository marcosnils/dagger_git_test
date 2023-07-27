package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-git/v5"

	"dagger.io/dagger"
)

func main() {
	server, err := NewGitServer()
	if err != nil {
		log.Fatal(err)
	}

	gitFiles, err := gitWorktreeFiles()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := server.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	defer server.Stop()

	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	gss := client.Host().UnixSocket("git-server.sock")

	cdir := client.Directory()
	if len(gitFiles) > 0 {
		cdir = client.Host().Directory(".", dagger.HostDirectoryOpts{
			Include: gitFiles,
		})
	}

	gitService := client.Container().From("alpine").
		WithUnixSocket("/git-server.sock", gss).
		WithExec([]string{"apk", "add", "socat"}).
		WithExec([]string{"socat", "TCP-LISTEN:9418,fork", "UNIX-CONNECT:/git-server.sock"}).
		WithExposedPort(9418)

	hostname, err := gitService.Hostname(ctx)
	if err != nil {
		log.Fatal(err)
	}

	gitDir := client.Git("git://"+hostname+"/.git", dagger.GitOpts{
		ExperimentalServiceHost: gitService,
		KeepGitDir:              true,
	}).Branch("main").Tree()

	_, err = client.Container().From("golang:1.20").
		WithDirectory("/app", gitDir).
		WithDirectory("/app", cdir).
		WithWorkdir("/app").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"go", "build", "-o", "dagger_git_test"}).
		File("dagger_git_test").Export(ctx, "dagger_git_test")

	fmt.Println(err)
}

func gitWorktreeFiles() ([]string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	ret := []string{}
	st, err := wt.Status()
	if err != nil {
		return nil, err
	}
	for k := range st {
		ret = append(ret, k)
	}
	return ret, nil
}
