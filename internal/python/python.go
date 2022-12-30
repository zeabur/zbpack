package python

import (
	"os"

	"github.com/zeabur/zbpack/pkg/types"
)

func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	framework := meta["framework"]
	entry := meta["entry"]
	dependencyPolicy := meta["dependencyPolicy"]

	dockerfile := "FROM python:3.8.2-slim-buster\n"

	installCmds := ""

	switch dependencyPolicy {
	case "requirements.txt":
		installCmds = `ADD requirements.txt requirements.txt
RUN pip install -r requirements.txt`
	case "poetry.lock":
		installCmds = `ADD pyproject.toml pyproject.toml
ADD poetry.lock poetry.lock
RUN pip install poetry
RUN poetry config virtualenvs.create false
RUN poetry install --no-dev`
	case "pyproject.toml":
		installCmds = `ADD pyproject.toml pyproject.toml
RUN pip install poetry
RUN poetry config virtualenvs.create false
RUN poetry install --no-dev`
	case "Pipfile":
		installCmds = `ADD Pipfile Pipfile
ADD Pipfile.lock Pipfile.lock
RUN pip install pipenv
RUN pipenv install --system --deploy --ignore-pipfile`
	}

	if framework == string(types.PythonFrameworkDjango) {
		dir, err := os.ReadDir("/src")
		if err != nil {
			return "", err
		}
		for _, d := range dir {
			if d.IsDir() {
				if _, err := os.Stat("/src/" + d.Name() + "/wsgi.py"); err == nil {
					entry = d.Name() + ".wsgi"
				}
			}
		}

		dockerfile += `WORKDIR /app
` + installCmds + `
COPY . .
EXPOSE 8080
RUN pip install gunicorn
CMD gunicorn --bind :8080 ` + entry

	} else {
		dockerfile += `WORKDIR /app
COPY . .
` + installCmds + `
EXPOSE 8080
CMD python ` + entry

	}

	return dockerfile, nil
}
