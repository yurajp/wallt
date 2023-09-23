package main

import (
    "fmt"
    "net/http"
    "html/template"
    "os/exec"
    "time"
    "context"
  
    _ "github.com/mattn/go-sqlite3"
)


func makeTempls() (map[string]*template.Template, error) {
  templs := map[string]*template.Template{}
  routes := []string{"wellcome",
  "home", "allSites", "allCards", "oneSite", "oneCard", "createSite",
  "createCard", "allDocs", "createDoc", "passrf", "createPassrf", "message"}
  for _, r := range routes {
    t, err := template.ParseFiles("templates/" + r + ".html")
    if err != nil {
      return templs, err
    }
    templs[r] = t
  }
  return templs, nil
}

func TimeControl(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if app.IsDead() {
      http.Redirect(w, r, "/", 302)
      return
    }
    if app.IsBusy() {
      <-app.web.trans
    }
    next.ServeHTTP(w, r)
    
    go func() {
      if !app.IsBusy() {
        app.web.trans <-struct{}{}
      }
      select {
        case <-time.After(livetime):
          app.Dies()
          fmt.Println("\t Time over!")
          return
        case app.web.trans<-struct{}{}:
          fmt.Print(" ✓")
          return
      }
    }()
  })
}

func NewWeb() *Web{
    mux := http.NewServeMux()
    fs := http.FileServer(http.Dir("./static"))
    mux.Handle("/static/", http.StripPrefix("/static/", fs))
    muxMap := map[string]func(http.ResponseWriter, *http.Request){"/": wellcome,
      "/createSite": createSiteWeb,
      "/createCard": createCardWeb,
      "/createDoc": createDocWeb,
      "/createPassrf": createPassrfWeb,
      "/backup": BackupWeb,
      "/recode": RecodeWeb,
      "/share": ShareWeb,
      "/exit": exitWeb,
      "/quit": quitApp,
    }
    for p, fn := range muxMap {
      mux.HandleFunc(p, fn)
    }
    midMap:= map[string]func(http.ResponseWriter, *http.Request){"/home": homeHandler,
     "/sites": allSitesWeb,
     "/cards": allCardsWeb,
     "/site": showSiteWeb,
     "/card": showCardWeb,
     "/docs": showDocsWeb,
     "/deleteSite": deleteSiteWeb,
     "/deleteCard": deleteCardWeb,
     "/passrf": showPassrfWeb,
    }
    for p, fn := range midMap {
      mux.Handle(p, TimeControl(http.HandlerFunc(fn)))
    }
    server := &http.Server{
      Addr: port,
      Handler: mux,
    }
    templs, err := makeTempls()
    if err != nil {
      fmt.Println(err)
      return &Web{}
    }
    ctx := context.Background()
    trans := make(chan struct{}, 1)
    web := Web{server, ctx, templs, "", trans}
    return &web
}

func (app *App) Run() {
    go func() {
      defer close(app.web.trans)
        err := app.web.server.ListenAndServe()
        check(err)
    }()
    fmt.Printf("\t Safe time: %v min\n", livetime.Minutes())
    cmd := exec.Command("termux-open-url", fmt.Sprintf("http://localhost%s/", port))
    err := cmd.Run()
    check(err)
  //  var q string
  //  fmt.Println("\n\t WALLT RUNS\n\t Enter any to quit")
   // fmt.Scanf("%s", q)
}