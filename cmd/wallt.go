package main

import (
  "fmt"
  "database/sql"
  "os/exec"
  "context"
  "time"
  
  "github.com/yurajp/wallt/conf"
  "github.com/yurajp/wallt/internal/web"
  "github.com/yurajp/wallt/internal/database"
)


type App struct {
  web *web.Web
  wdb *database.Wdb
}

var (
  app *App
  port string
  livetime time.Duration
)


func NewApp(w *web.Web, d *database.Wdb) *App {
  return &App{w, d}
}

func (app *App) Run() {
    go func() {
      defer close(app.web.Trans)
      err := app.web.Server.ListenAndServe()
      if err != nil {
        fmt.Println(" SERVER ERROR: ", err)
        app.web.Quit <-struct{}{}
          return
      }
    }()
    
    prt := app.web.Server.Addr
    fmt.Printf("\n\n\t   WALLT\n\n  Port %s Safetime: %v min\n\n", prt, livetime.Minutes())
    cmd := exec.Command("termux-open-url", fmt.Sprintf("http://localhost%s/", prt))
    err := cmd.Run()
    if err != nil {
      fmt.Println(err)
    }
}


func main() {
    err := conf.Prepare()
    if err != nil {
      fmt.Println(err)
      return
    }
    err = conf.GetConfig()
    if err != nil {
      fmt.Println(err)
      return
    }
    port = conf.Cfg.Port
    livetime = conf.Cfg.Livetime
    
    db, err := sql.Open("sqlite3", "data/wallt.db")
    if err != nil {
        fmt.Println(err)
       return
    }
    defer db.Close()
    defer func() {
      clpb := exec.Command("termux-clipboard-set", " ")
      err := clpb.Run()
      if err != nil {
        fmt.Println(err)
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
