package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

const MAX_UPLOAD_SIZE = 1024 * 1024 * 1024 * 10 // 10GB

func RandBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return []byte("0"), err
	}
	return b, nil
}

func timeHandler(format string) http.Handler {
	fmt.Printf("request /time")
	fn := func(w http.ResponseWriter, r *http.Request) {
		tm := time.Now().Format(format)
		w.Write([]byte("" + tm))
	}
	return http.HandlerFunc(fn)
}

func dummyFileHandler() http.Handler {
	fmt.Printf("request /file")
	fn := func(w http.ResponseWriter, r *http.Request) {

		// size in megabyte, default is 100
		param_size := r.URL.Query().Get("size")
		var size int64 = 100

		if temp_size, err := strconv.ParseInt(param_size, 10, 64); err == nil {
			size = temp_size * 1024 * 1024
		}

		fmt.Printf("Serving file with size: %v MB .\n", size/1024/1024)

		block_size := 4096
		someRandomBytes, _ := RandBytes(block_size)

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)

		var count_bytes int64 = 0

		for ; count_bytes < size; count_bytes += int64(block_size) {
			if math.Mod(float64(count_bytes), float64(1024*1024*100)) == 0 {
				fmt.Printf("send block ... %v \n", count_bytes/1024/1024)
			}
			w.Write(someRandomBytes)
		}
		fmt.Printf("finished: %v MB.\n", count_bytes/1024/1024)
	}
	return http.HandlerFunc(fn)
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// https://mustafanafizdurukan.github.io/posts/understanding-io-teereader/
type progress struct {
	total uint64
}

func (p *progress) Write(b []byte) (int, error) {
	p.total += uint64(len(b))
	fmt.Printf("got %d bytes...\n", p.total)
	return len(b), nil
}

type LogProgressWriter struct{}

func (pw LogProgressWriter) Write(data []byte) (int, error) {
	// implement progress here
	fmt.Printf("wrote %d bytes\n", len(data))
	return len(data), nil
}

// https://github.com/Freshman-tech/file-upload
func dummyUploadHandler() http.Handler {
	fmt.Printf("request /upload")
	fn := func(w http.ResponseWriter, r *http.Request) {

		// we do not allow PUT, DELETE,.. because thus does not support MultiPart
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 32 MB is the default used by FormFile
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// get a reference to the fileHeaders
		files := r.MultipartForm.File["file"]

		start := time.Now()

		for _, fileHeader := range files {
			if fileHeader.Size > MAX_UPLOAD_SIZE {
				http.Error(w, fmt.Sprintf("The uploaded image is too big: %s. Please use an image less than 10GB in size", fileHeader.Filename), http.StatusBadRequest)
				return
			}

			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer file.Close()

			buff := make([]byte, 512)
			_, err = file.Read(buff)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = file.Seek(0, io.SeekStart)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			progressReader := io.TeeReader(file, LogProgressWriter{})

			_, err = io.Copy(io.Discard, progressReader)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

		}

		elapsed := time.Since(start)

		fmt.Fprintf(w, "Upload successful, took: %ds, ", elapsed.Milliseconds()/1000)
	}

	return http.HandlerFunc(fn)
}

func main() {
	DUMMY_SERVER_BIND := getenv("DUMMY_SERVER_BIND", ":8000")

	rh := http.RedirectHandler("about:blank", 307)
	http.Handle("/", rh)

	th := timeHandler(time.RFC3339Nano)
	http.Handle("/time", th)

	dfh := dummyFileHandler()
	http.Handle("/file", dfh)

	duh := dummyUploadHandler()
	http.Handle("/upload", duh)

	fmt.Printf("listening: " + DUMMY_SERVER_BIND)

	err := http.ListenAndServe(DUMMY_SERVER_BIND, nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
