package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type ConfigRepository struct {
	Repository       string `yaml:"repository"`
	TemplateFileName string `yaml:"templateFileName"`
}

type Config struct {
	Repositories []*ConfigRepository `yaml:"repositories"`
}

type ReleaseShortView struct {
	CreatedAt        time.Time `json:"createdAt"`
	IsDraft          bool      `json:"isDraft"`
	IsLatest         bool      `json:"isLatest"`
	IsPrerelease     bool      `json:"isPrerelease"`
	Name             string    `json:"name"`
	PublishedAt      time.Time `json:"publishedAt"`
	TagName          string    `json:"tagName"`
	TarballUrl       string    `json:"tarballUrl"`
	TarballUrlSHA256 string    `json:"tarballUrlSha256"`
	Prefix           string    `json:"prefix"`
}

type ReleaseView struct {
	TarballUrl string `json:"tarballUrl"`
}

func GetUrlFileHashSHA256(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	slice, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	if _, err := hash.Write(slice); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func main() {
	regExp := regexp.MustCompile("^(-)|[^0-9]+")

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	byteSlice, err := os.ReadFile("./util/config/config.yaml")
	if err != nil {
		logger.Fatal(
			"Error reading config.json",
			zap.Error(err),
		)
	}

	var config Config
	if err = yaml.Unmarshal(byteSlice, &config); err != nil {
		logger.Fatal(
			"Error parsing config.json",
			zap.Error(err),
		)
	}

	logger.Info("loaded config.yaml", zap.Any("config", config))

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		go func() {
			defer cancel()

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

			select {
			case <-ctx.Done():
			case <-quit:
			}
		}()
	}()

	for _, repo := range config.Repositories {
		split := strings.Split(repo.Repository, "/")
		configFileName := split[len(split)-1]

		tmpl, err := template.ParseFiles(repo.TemplateFileName)
		if err != nil {
			logger.Fatal(
				"Error parsing template",
				zap.Error(err),
			)
		}

		logger.Info("loaded template", zap.Any("template", tmpl.Name()))

		logger.Info("starting repository", zap.String("link", repo.Repository))

		var stdoutShortViews, stderrShortViews = bytes.NewBuffer(nil), bytes.NewBuffer(nil)
		cmdShortViews := exec.CommandContext(
			ctx,
			"gh",
			"release",
			"list",
			"--json",
			"createdAt,isDraft,isLatest,isPrerelease,name,publishedAt,tagName",
			"-R",
			repo.Repository,
		)

		cmdShortViews.Stderr = stderrShortViews
		cmdShortViews.Stdout = stdoutShortViews

		if err = cmdShortViews.Run(); err != nil {
			logger.Fatal(
				"Error running command",
				zap.Error(err),
				zap.String("stderr", stderrShortViews.String()),
			)
		}

		logger.Info(
			"get release info",
			zap.String("stdout", stdoutShortViews.String()),
		)

		var releaseViews []*ReleaseShortView
		if err = json.Unmarshal(stdoutShortViews.Bytes(), &releaseViews); err != nil {
			logger.Fatal(
				"Error parsing release views",
				zap.Error(err),
			)
		} else if len(releaseViews) == 0 {
			logger.Fatal(
				"no release views found",
				zap.String("stderr", stderrShortViews.String()),
				zap.String("stdout", stdoutShortViews.String()),
			)
		}

		logger.Info(
			"get release views",
			zap.Int("count of releases", len(releaseViews)),
			zap.Any("releases", releaseViews),
		)

		var latest *ReleaseShortView
		for _, releaseView := range releaseViews {
			if releaseView.IsLatest {
				latest = releaseView
			}

		}

		if latest == nil {
			logger.Fatal("no latest release found")
		}

		releaseViewsUpdateSyncGroup := &sync.WaitGroup{}
		releaseViewsUpdateSyncGroup.Add(len(releaseViews))
		for _, view := range releaseViews {
			go func(
				view *ReleaseShortView,
				group *sync.WaitGroup,
			) {
				defer group.Done()

				logger.Info("get view of release", zap.Any("view", view))

				var stdout, stderr = bytes.NewBuffer(nil), bytes.NewBuffer(nil)
				cmdView := exec.CommandContext(
					ctx,
					"gh",
					"release",
					"view",
					view.Name,
					"--json",
					"tarballUrl",
					"-R",
					repo.Repository,
				)

				cmdView.Stderr = stderr
				cmdView.Stdout = stdout

				if err = cmdView.Run(); err != nil {
					logger.Fatal(
						"can not get release view",
						zap.Error(err),
						zap.String("stderr", stderr.String()),
						zap.String("stdout", stdout.String()),
						zap.String("view", view.Name),
					)
				}

				logger.Info("get view of release", zap.Any("view", view))

				var modelView *ReleaseView
				if err = json.Unmarshal(stdout.Bytes(), &modelView); err != nil {
					logger.Fatal(
						"can not unmarshal release view",
						zap.Error(err),
						zap.String("stderr", stderr.String()),
						zap.String("stdout", stdout.String()),
						zap.String("view", view.Name),
					)
				}

				logger.Info("get TarballUrl", zap.Any("TarballUrl", modelView.TarballUrl))

				if modelView.TarballUrl == "" {
					logger.Fatal(
						"TarballUrl is empty",
						zap.String("stderr", stderr.String()),
						zap.String("stdout", stdout.String()),
						zap.String("view", view.Name),
					)
				}

				view.TarballUrl = modelView.TarballUrl

				if view.TarballUrlSHA256, err = GetUrlFileHashSHA256(modelView.TarballUrl); err != nil {
					logger.Fatal(
						"can not get sha256 TarballUrl",
						zap.Error(err),
						zap.String("stderr", stderr.String()),
						zap.String("stdout", stdout.String()),
						zap.String("view", view.Name),
					)
				}

				logger.Info("loaded view of release", zap.Any("view", view))

				buf := bytes.NewBuffer(nil)
				if err := tmpl.Execute(buf, view); err != nil {
					logger.Fatal(
						"Error executing template",
						zap.Error(err),
					)
				}

				fileName := "./Formula/" + configFileName + "@" + regExp.ReplaceAllString(view.TagName, "") + ".rb"
				WriteFile(logger, fileName, buf.Bytes())

				if view.IsLatest {
					fileName := "./Formula/" + configFileName + ".rb"
					WriteFile(logger, fileName, buf.Bytes())
				}
			}(view, releaseViewsUpdateSyncGroup)
		}
		releaseViewsUpdateSyncGroup.Wait()
	}
}

func WriteFile(
	logger *zap.Logger,
	fileName string,
	data []byte,
) {
	file, err := os.Create(fileName)
	if err != nil {
		logger.Fatal(
			"Error creating file",
			zap.Error(err),
			zap.String("fileName", fileName),
		)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Fatal(
				"Error closing file",
				zap.Error(err),
				zap.String("fileName", fileName),
			)
		}
	}()

	if _, err := file.Write(data); err != nil {
		logger.Fatal(
			"Error writing file",
			zap.Error(err),
			zap.String("fileName", fileName),
		)
	}

	logger.Info("write file", zap.String("fileName", fileName))
}
