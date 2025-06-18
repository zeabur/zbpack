package zeaburpack

import (
	"strings"
	"testing"

	"github.com/zeabur/zbpack/pkg/types"
)

func TestInjectLabels(t *testing.T) {
	tests := []struct {
		name          string
		dockerfile    string
		planType      types.PlanType
		planMeta      types.PlanMeta
		expectedLines []string
	}{
		{
			name: "Inject labels after FROM statement",
			dockerfile: `FROM node:18-alpine
RUN mkdir /app
WORKDIR /app`,
			planType: types.PlanTypeNodejs,
			planMeta: types.PlanMeta{
				"framework": "next.js",
			},
			expectedLines: []string{
				"FROM node:18-alpine",
				`LABEL "language"="nodejs"`,
				`LABEL "framework"="next.js"`,
				"RUN mkdir /app",
				"WORKDIR /app",
			},
		},
		{
			name: "Inject only language label when no framework",
			dockerfile: `FROM golang:1.20-alpine
WORKDIR /src`,
			planType: types.PlanTypeGo,
			planMeta: types.PlanMeta{},
			expectedLines: []string{
				"FROM golang:1.20-alpine",
				`LABEL "language"="go"`,
				"WORKDIR /src",
			},
		},
		{
			name: "Inject labels with multi-stage dockerfile",
			dockerfile: `FROM golang:1.20-alpine AS builder
WORKDIR /src
FROM alpine AS runtime
COPY --from=builder /app /app`,
			planType: types.PlanTypeGo,
			planMeta: types.PlanMeta{
				"framework": "gin",
			},
			expectedLines: []string{
				"FROM golang:1.20-alpine AS builder",
				`LABEL "language"="go"`,
				`LABEL "framework"="gin"`,
				"WORKDIR /src",
				"FROM alpine AS runtime",
				"COPY --from=builder /app /app",
			},
		},
		{
			name: "Inject labels at beginning when no FROM statement",
			dockerfile: `RUN echo "test"
CMD ["echo", "hello"]`,
			planType: types.PlanTypePython,
			planMeta: types.PlanMeta{
				"framework": "fastapi",
			},
			expectedLines: []string{
				`LABEL "language"="python"`,
				`LABEL "framework"="fastapi"`,
				`RUN echo "test"`,
				`CMD ["echo", "hello"]`,
			},
		},
		{
			name: "Empty framework should not add framework label",
			dockerfile: `FROM python:3.9-slim
WORKDIR /app`,
			planType: types.PlanTypePython,
			planMeta: types.PlanMeta{
				"framework": "",
			},
			expectedLines: []string{
				"FROM python:3.9-slim",
				`LABEL "language"="python"`,
				"WORKDIR /app",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InjectLabels(tt.dockerfile, tt.planType, tt.planMeta)
			resultLines := strings.Split(result, "\n")

			if len(resultLines) != len(tt.expectedLines) {
				t.Errorf("Expected %d lines, got %d lines", len(tt.expectedLines), len(resultLines))
				t.Errorf("Expected:\n%s", strings.Join(tt.expectedLines, "\n"))
				t.Errorf("Got:\n%s", result)
				return
			}

			for i, expectedLine := range tt.expectedLines {
				if resultLines[i] != expectedLine {
					t.Errorf("Line %d: expected %q, got %q", i, expectedLine, resultLines[i])
				}
			}
		})
	}
}

func TestGenerateDockerfile(t *testing.T) {
	// Test that generateDockerfile calls injectLabels correctly
	opt := &GenerateDockerfileOptions{
		PlanType: types.PlanTypeStatic,
		PlanMeta: types.PlanMeta{
			"framework": "hugo",
		},
	}

	dockerfile, err := GenerateDockerfile(opt)
	if err != nil {
		t.Fatalf("generateDockerfile failed: %v", err)
	}

	// Verify that labels are injected
	if !strings.Contains(dockerfile, `LABEL "language"="static"`) {
		t.Error("Language label not found in generated Dockerfile")
	}

	if !strings.Contains(dockerfile, `LABEL "framework"="hugo"`) {
		t.Error("Framework label not found in generated Dockerfile")
	}
}
