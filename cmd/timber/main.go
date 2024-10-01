package main

import (
	"connectrpc.com/connect"
	"errors"
	"github.com/chushi-io/timber/gen/server/v1/serverv1connect"
	"github.com/chushi-io/timber/interceptor"
	"github.com/chushi-io/timber/internal/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	mux.Handle(path, handler)
	mux.HandleFunc("/ping", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /files/{file}", func(writer http.ResponseWriter, request *http.Request) {
		file := request.PathValue("file")
		filePath := filepath.Join(logDir, file)
		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		// optionally, filter file contents
		contents, err := os.ReadFile(filePath)
		if err != nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		writer.Write(contents)
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
