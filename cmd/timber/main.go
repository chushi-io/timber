package main

import (
	"github.com/chushi-io/timber/gen/server/v1/serverv1connect"
	"github.com/chushi-io/timber/internal/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log"
	"net/http"
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
	srv := server.New()
	mux := http.NewServeMux()
	path, handler := serverv1connect.NewLogsServiceHandler(srv)
	mux.Handle(path, handler)
	mux.HandleFunc("/ping", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("OK"))
	})

	http.ListenAndServe(
		address,
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	)
}
