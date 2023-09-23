package main

import (
  "github.com/yurajp/wallt/conf"
  "fmt"
  "database/sql"
  "os/exec"
  "context"
  "time"
)


var (
  app *App
  port string
  livetime time.Duration
  chanq = make(chan struct{})
)

func main() {
    err := conf.Prepare()
    if err != nil {
      fmt.Println(err)
      return
    }
    cfg, err := conf.GetConfig()
    if err != nil {
      fmt.Println(err)
      return
    }
    port = cfg.Port
    livetime = cfg.Livetime
    db, err := sql.Open("sqlite3", "wallt.db")
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
    web := NewWeb()
    app = &App{web, db}
    ctx, cancel := context.WithCancel(app.web.ctx)
    defer cancel()
    app.web.ctx = ctx
    
    go app.Run()
    <-chanq
}
