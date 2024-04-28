package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
	"github.com/raihan88/snippetbox/pkg/models/mysql"
)


type contextKey string
var contextKeyUser = contextKey("user")


type application struct{
	errorLog *log.Logger
	infoLog *log.Logger
	session *sessions.Session
	snippets *mysql.SnippetModel
	templateCache map[string]*template.Template
	users *mysql.UserModel
	
}


func main(){

	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:Ra!han88@/snippetbox?parseTime=true", "MySQL Data Source Name")
	flag.Parse()

	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret Key For Session Managing")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate | log.LUTC)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db,err := openDB(*dsn)
	if err != nil{
		errorLog.Fatal(err)
	}

	defer db.Close()

	templateCache , err := newTemplateCache("ui/html/")
	if err != nil{
		errorLog.Fatal(err)
	}

	session := sessions.New([]byte(*secret))
	session.Lifetime = 1 * time.Minute
	session.Secure = true
	session.SameSite = http.SameSiteStrictMode


	app := &application{
		errorLog: errorLog,
		infoLog: infoLog,
		session: session,
		snippets: &mysql.SnippetModel{DB:db},
		templateCache: templateCache,
		users: &mysql.UserModel{DB:db},
	}

	
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}


	srv := &http.Server{
		Addr: *addr,
		ErrorLog: errorLog,
		Handler: app.routes(),
		TLSConfig: tlsConfig,
		IdleTimeout: time.Minute,
		ReadTimeout: 5*time.Second,
		WriteTimeout: 10*time.Second,
	}

	infoLog.Printf("Starting a server on: %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)

}


func openDB(dsn string)(*sql.DB, error){
	db,err := sql.Open("mysql", dsn)
	if err != nil{
		return nil, err 
	}

	if err = db.Ping(); err != nil{
		return nil, err 
	}
	return db, nil
}