package gitlab

import (
	"regexp"
	"strings"
)

type MergeRequest struct {
	IID  int64
	Refs MergeRequestRefs
}

type MergeRequestRefs struct {
	BaseSHA  string
	HeadSHA  string
	StartSHA string
}

type MergeRequestDiff struct {
	Path string
	Diff string
}

var filePathRegex = regexp.MustCompile(`(?m)^\+\+\+ b/(.+)$`)

func ParseMergeRequestDiff(rawBytes []byte) []*MergeRequestDiff {
	rawDiff := string(rawBytes)
	if rawDiff == "" {
		return nil
	}

	var result []*MergeRequestDiff

	for _, chunk := range strings.Split(rawDiff, "diff --git ") {
		if chunk == "" {
			continue
		}

		matches := filePathRegex.FindStringSubmatch(chunk)
		if len(matches) < 2 {
			continue
		}

		filePath := strings.TrimSpace(matches[1])

		result = append(result, &MergeRequestDiff{
			Path: filePath,
			Diff: "diff --git " + chunk,
		})
	}

	return result
}
