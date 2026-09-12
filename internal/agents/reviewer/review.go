package reviewer

import (
	"context"
	"embed"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/vladimirkorzhenevskiy/goops/pkg/gitlab"
	"github.com/vladimirkorzhenevskiy/goops/pkg/llm"
	"golang.org/x/sync/errgroup"
)

var ErrInvalidURL = errors.New("invalid gitlab merge request url")

//go:embed rules/*
var rulesFS embed.FS

type GitLab interface {
	GetMergeRequest(ctx context.Context, projectID string, mergeRequestID int64) (*gitlab.MergeRequest, error)
	FetchMergeRequestDiff(ctx context.Context, projectID string, mergeRequestID int64) ([]*gitlab.MergeRequestDiff, error)
	CreateMergeRequestDiscussion(
		ctx context.Context,
		projectID string,
		mergeRequestID int64,
		refs gitlab.MergeRequestRefs,
		filePath string,
		line int64,
		text string,
	) error
}

type LLM interface {
	Request(ctx context.Context, req llm.Request) (io.Reader, error)
}

type ReviewUseCase struct {
	gitLab GitLab
	llm    LLM
}

func NewReviewUseCase(gitLab GitLab, llm LLM) (*ReviewUseCase, error) {
	return &ReviewUseCase{
		llm:    llm,
		gitLab: gitLab,
	}, nil
}

func (uc *ReviewUseCase) Exec(ctx context.Context, mergeRequestURL string) error {
	projectID, mergeRequestID, err := parseMergeRequestURL(mergeRequestURL)
	if err != nil {
		return fmt.Errorf("parse merge request url: %w", err)
	}

	mergeRequest, err := uc.gitLab.GetMergeRequest(ctx, projectID, mergeRequestID)
	if err != nil {
		return fmt.Errorf("get merge request: %w", err)
	}

	diff, err := uc.gitLab.FetchMergeRequestDiff(ctx, projectID, mergeRequest.IID)
	if err != nil {
		return fmt.Errorf("fetch merge request diff: %w", err)
	}

	diff = slices.DeleteFunc(diff, func(item *gitlab.MergeRequestDiff) bool {
		if strings.HasSuffix(item.Path, "_gen.go") {
			return true
		}

		if strings.HasSuffix(item.Path, "_test.go") {
			return true
		}

		if strings.HasSuffix(item.Path, ".go") {
			return false
		}

		if strings.HasSuffix(item.Path, ".sql") {
			return false
		}

		return true
	})

	prompt, err := getSystemPrompt()
	if err != nil {
		return fmt.Errorf("get system prompt: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(5)

	for i := range diff {
		item := diff[i]

		// Если файл пустой (например, поменялись только права доступа), скипаем
		if item.Diff == "" {
			continue
		}

		g.Go(func() error {
			resp, err := uc.llm.Request(ctx, llm.Request{
				Messages: []llm.Message{
					llm.SystemMessage(prompt),
					llm.UserMessage(fmt.Sprintf("Файл: %s\n\n%s", item.Path, item.Diff)),
				},
			})
			if err != nil {
				return fmt.Errorf("llm request failed for %s: %w", item.Path, err)
			}

			data, err := io.ReadAll(resp)
			if err != nil {
				return fmt.Errorf("read response data: %w", err)
			}

			fmt.Printf("=== Ответ ИИ для файла: %s ===\n%s\n", item.Path, string(data))

			var res ReviewResponse

			if err := json.Unmarshal(data, &res); err != nil {
				return fmt.Errorf("unmarshal LLM response data: %w", err)
			}

			for _, issue := range res.Issues {
				if issue.Line <= 0 || issue.Comment == "" {
					continue
				}

				err := uc.gitLab.CreateMergeRequestDiscussion(
					ctx,
					projectID,
					mergeRequest.IID,
					mergeRequest.Refs,
					item.Path,
					issue.Line,
					fmt.Sprintf("### 🤖 AI Reviewer\n\n%s", issue.Comment),
				)
				if err != nil {
					fmt.Printf("❌ Не удалось опубликовать коммент в %s:%d: %v\n", item.Path, issue.Line, err)
				}
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("parallel review execution failed: %w", err)
	}

	return nil
}

func getSystemPrompt() (string, error) {
	data, err := rulesFS.ReadFile("rules/rules.md")
	if err != nil {
		return "", fmt.Errorf("read `rules/rules.md`: %w", err)
	}

	return string(data), nil
}

func parseMergeRequestURL(mergeRequestURL string) (string, int64, error) {
	regExp := regexp.MustCompile(`(?:/projects)?/([\w-]+/[\w-]+/[\w-]+)/-/merge_requests/(\d+)`)

	matches := regExp.FindStringSubmatch(strings.TrimSpace(mergeRequestURL))
	if len(matches) < 3 {
		return "", 0, ErrInvalidURL
	}

	projectID := matches[1]

	mrID, err := strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("parse mr id: %w", err)
	}

	return projectID, mrID, nil
}

func PrepareDiffForAI(rawDiff string) string {
	if rawDiff == "" {
		return ""
	}

	lines := strings.Split(rawDiff, "\n")
	var result []string

	for _, line := range lines {
		// Оставляем заголовки файлов и ханки (@@ -2,58 +2,38 @@)
		if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") || strings.HasPrefix(line, "@@ ") {
			result = append(result, line)
			continue
		}

		// Если строка добавлена — оставляем обязательно
		if strings.HasPrefix(line, "+") {
			result = append(result, line)
			continue
		}

		// Если строка удалена — мы заменяем её на супер-компактный маркер,
		// чтобы ИИ знал, что тут было удаление, но не читал 600 строк текста!
		if strings.HasPrefix(line, "-") {
			result = append(result, "-") // Оставляем просто минус без текста
			continue
		}

		// Обычные строки контекста оставляем как есть
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

type ReviewResponse struct {
	Issues []ReviewComment `json:"issues"`
}

type ReviewComment struct {
	Line    int64  `json:"line"`
	Comment string `json:"comment"`
}
