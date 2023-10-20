package main

import (
  "fmt"
  "database/sql"
  "os/exec"
  "context"
  "time"
  "os"
  
  "github.com/yurajp/wallt/conf"
  "github.com/yurajp/wallt/internal/web"
  "github.com/yurajp/wallt/internal/database"
	"github.com/sirupsen/logrus"
)


type App struct {
  web *web.Web
  wdb *database.Wdb
}

var (
  app *App
  port string
  livetime time.Duration
  log *logrus.Logger
)


func NewApp(w *web.Web, d *database.Wdb) *App {
  return &App{w, d}
}

func (app *App) Run() {
    go func() {
      defer close(app.web.Trans)
      err := app.web.Server.ListenAndServe()
      if err != nil {
        log.WithError(err).Error("Cannot start server")
        app.web.Quit <-struct{}{}
          return
      }
    }()
    
    prt := app.web.Server.Addr
    log.Infof("WALLT\n\n  runs on %s safetime: %v min\n\n", prt, livetime.Minutes())
    cmd := exec.Command("termux-open-url", fmt.Sprintf("http://localhost%s/", prt))
    err := cmd.Run()
    if err != nil {
      log.WithError(err).Errorf("Cannot open localhost%s", prt)
    }
}


func main() {
	  log = logrus.New()
	  log.Formatter = new(logrus.TextFormatter)
	  log.Level = logrus.InfoLevel
	  log.Out = os.Stdout
    err := conf.Prepare()
    if err != nil {
      log.WithError(err).Error("Cannot prepare config")
    }
    err = conf.GetConfig()
    if err != nil {
      log.WithError(err).Error("Cannot get Config")
    }
    port = conf.Cfg.Port
    livetime = conf.Cfg.Livetime
    
    db, err := sql.Open("sqlite3", "../data/wallt.db")
    if err != nil {
        log.WithError(err).Error("Cannot open DB")
    }
    defer db.Close()
    defer func() {
      clpb := exec.Command("termux-clipboard-set", " ")
      err := clpb.Run()
      if err != nil {
        log.WithError(err).Warn("Cannot set clipboard")
      }
    }()
    
    web.WEB = web.NewWeb()
    database.WDB = database.NewWdb(db)
    
    app = NewApp(web.WEB, database.WDB)
    ctx, cancel := context.WithCancel(app.web.Ctx)
    defer cancel()
    app.web.Ctx = ctx
    
    go app.Run()
    
    Work:
    for {
      select {
        case <-web.WEB.Quit:
          break Work
        default:
      }
    }
}
