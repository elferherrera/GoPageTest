package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"fjherrera.net/snippetbox/pkg/models/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
)

type application struct {
	logInfo       *log.Logger
	logError      *log.Logger
	session       *sessions.Session
	snippets      *mysql.SnippetModel
	templateCache map[string]*template.Template
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	staticFolder := flag.String("static", "./ui/static/", "Folder with static files")
	dsn := flag.String("dsn", "root:secret@tcp(localhost:33060)/snippetbox?parseTime=true", "MySQL data")
	secret := flag.String("secret", "supersecret", "Secret value")
	flag.Parse()

	logInfo := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	logError := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dsn)
	if err != nil {
		logError.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		logError.Fatal(err)
	}

	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour

	app := &application{
		logInfo:       logInfo,
		logError:      logError,
		session:       session,
		snippets:      &mysql.SnippetModel{DB: db},
		templateCache: templateCache,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: logError,
		Handler:  app.routes(*staticFolder),
	}

	logInfo.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	logError.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
