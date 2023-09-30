package web

import (
    "fmt"
    "net/http"
    "html/template"
    "time"
    "context"
  
    _ "github.com/mattn/go-sqlite3"
    "github.com/yurajp/wallt/conf"
    "github.com/yurajp/wallt/internal/models"
)

type Web models.Web

var (
  WEB *Web
  Livetime = conf.Cfg.Livetime
  Port = conf.Cfg.Port
)

func makeTempls() (map[string]*template.Template, error) {
  templs := map[string]*template.Template{}
  routes := []string{"wellcome",
  "home", "allSites", "allCards", "oneSite", "oneCard", "createSite",
  "createCard", "allDocs", "createDoc", "passrf", "createPassrf", "message", "editDoc"}
  tdir := conf.Cfg.Appdir + "/internal/web/templates/"
  for _, r := range routes {
    t, err := template.ParseFiles(tdir + r + ".html")
    if err != nil {
      return templs, err
    }
    templs[r] = t
  }
  return templs, nil
}

func TimeControl(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if WEB.IsDead() {
      fmt.Println("\n   Oops! Web is DEAD!")
      http.Redirect(w, r, "/", 302)
      return
    }
    if WEB.IsBusy() {
      <-WEB.Trans
    }
    next.ServeHTTP(w, r)
    
    go func() {
      if !WEB.IsBusy() {
        WEB.Trans <-struct{}{}
      }
      lt := conf.Cfg.Livetime
      select {
        case <-time.After(lt):
          WEB.Dies()
          fmt.Println("\t Time over!")
          return
        case WEB.Trans<-struct{}{}:
          fmt.Print(" ✓")
          return
      }
    }()
  })
}

func NewWeb() *Web {
    fsdir := fmt.Sprintf("%s/internal/web/static", conf.Cfg.Appdir)
    mux := http.NewServeMux()
    fs := http.FileServer(http.Dir(fsdir))
    mux.Handle("/static/", http.StripPrefix("/static/", fs))
    muxMap := map[string]func(http.ResponseWriter, *http.Request){"/": wellcome,
      "/createSite": createSiteWeb,
      "/createCard": createCardWeb,
      "/createDoc": createDocWeb,
      "/editDoc": editDocWeb,
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
      Addr: conf.Cfg.Port,
      Handler: mux,
    }
    templs, err := makeTempls()
    if err != nil {
      fmt.Println(err)
      return &Web{}
    }
    ctx := context.Background()
    Trans := make(chan struct{}, 1)
    Quit := make(chan struct{}, 1)
    
    return &Web{server, ctx, templs, "", Trans, Quit}
}
