package main

import (
	"bytes"
	"connectrpc.com/connect"
	"errors"
	"fmt"
	"github.com/chushi-io/timber/gen/server/v1/serverv1connect"
	"github.com/chushi-io/timber/interceptor"
	"github.com/chushi-io/timber/internal/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var (
	logger *zap.Logger
)

var rootCmd = &cobra.Command{
	Use:   "timber",
	Short: "Run the timber server",
	Run:   runServer,
}

func init() {
	rootCmd.Flags().String("address", ":8080", "Address to bind to")
	rootCmd.Flags().String("log-dir", "/timber/data", "Directory to store logs in")
	rootCmd.Flags().Bool("debug", false, "Enable debug logging")
	logger, _ = zap.NewProduction()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
func logMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func runServer(cmd *cobra.Command, args []string) {
	//conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	//	os.Exit(1)
	//}
	//defer conn.Close(context.Background())

	address, _ := cmd.Flags().GetString("address")
	logDir, _ := cmd.Flags().GetString("log-dir")
	debug, _ := cmd.Flags().GetBool("debug")
	if debug {
		logger, _ = zap.NewDevelopment()
	}
	srv := server.New(logDir, logger)
	mux := http.NewServeMux()
	interceptors := connect.WithInterceptors(interceptor.NewServerAuthInterceptor())
	path, handler := serverv1connect.NewLogsServiceHandler(
		srv,
		interceptors,
	)

	mux.Handle(path, logMw(handler))

	mux.HandleFunc("GET /files/{file}", func(writer http.ResponseWriter, request *http.Request) {
		limit := request.URL.Query().Get("limit")
		offset := request.URL.Query().Get("offset")

		file := request.PathValue("file")
		fmt.Printf("Checking file %s\n", file)
		filePath := filepath.Join(logDir, file)
		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			fmt.Println("File doesnt exist")
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		logFile, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
		if err != nil {
			fmt.Println("Failed opening file")
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		// optionally, filter file contents
		if limit == "" && offset == "" {
			var buffer bytes.Buffer
			io.Copy(&buffer, logFile)
			if err != nil {
				fmt.Println("Failed copying buffer")
				writer.WriteHeader(http.StatusNoContent)
				return
			}
			writer.Write(buffer.Bytes())
		} else {
			intLimit, _ := strconv.Atoi(limit)
			intOffset, _ := strconv.Atoi(offset)
			_, err = logFile.Seek(int64(intOffset), 0)
			if err != nil {
				fmt.Println("Failed seeking file")
				writer.WriteHeader(http.StatusNoContent)
				return
			}
			out := make([]byte, intLimit)
			_, err = logFile.Read(out)
			if err != nil {
				fmt.Println("Failed reading requested bytes")
				fmt.Println(err)
				writer.WriteHeader(http.StatusNoContent)
				return
			}
			writer.Write(out)
		}
	})

	mux.HandleFunc("DELETE /files/{file}", func(writer http.ResponseWriter, request *http.Request) {
		file := request.PathValue("file")
		filePath := filepath.Join(logDir, file)
		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			// path/to/whatever does not exist
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		if err := os.Remove(filePath); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

	logger.Info("Starting server")
	http.ListenAndServe(
		address,
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	)
}
