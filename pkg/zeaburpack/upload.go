package zeaburpack

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseMultipartForm(32 << 20) // Set maxMemory to 32MB
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		relativePath := r.FormValue("relative_path")
		if relativePath == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		savePath := filepath.Join(".zeabur/output/static", relativePath)
		saveDir := filepath.Dir(savePath)

		if _, err := os.Stat(saveDir); os.IsNotExist(err) {
			err = os.MkdirAll(saveDir, 0755)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		out, err := os.Create(savePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer out.Close()

		if _, err = io.Copy(out, file); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}
