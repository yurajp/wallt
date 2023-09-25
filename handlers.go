package main

import (
    "fmt"
    "net/http"
    "sort"
    "strings"
    "html/template"
  
    "github.com/yurajp/wallt/purecrypt"
)

type Export struct {
  Exists bool
  Port string
}

func wellcome(w http.ResponseWriter, r *http.Request) {
  exists := purecrypt.ChWordExists()
  if r.Method == http.MethodGet {
    if wc, ok := app.web.templs["wellcome"]; ok {
      exp := Export{exists, port}
      wc.Execute(w, exp)
    }
  }
  if r.Method == http.MethodPost {
    if !exists {
      err := app.createTables()
      check(err)
      err = r.ParseForm()
      check(err)
      word1 := r.FormValue("word1")
      word2 := r.FormValue("word2")
      if len(word1) < 5 || word1 != word2 {
        http.Redirect(w, r, "/", 302)
        return
      }
      err = purecrypt.WriteCheckword(word1)
      check(err)
      app.web.word = word1
      http.Redirect(w, r, "/home", 302)
    } else {
      err := r.ParseForm()
      check(err)
      word := r.FormValue("word")
      if purecrypt.IsCorrect(word) {
        app.web.word = word
        http.Redirect(w, r, "/home", 303)
      } else {
        fmt.Println("Wrong Password")
        http.Redirect(w, r, "/", 302)
      }
    }
  }
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    return
  }
  err := app.execTempl(w, "home", port)
  if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func exitWeb(w http.ResponseWriter, r *http.Request) {
  app.Dies()
  http.Redirect(w, r, "/", 302)
  if len(app.web.trans) == 1 {
    <-app.web.trans
  }
}

func allSitesWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    return
  }
  list, err := app.GetAllSitesFromDb()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  names := []template.HTML{}
  sort.Slice(list, func(i, j int) bool { 
    return strings.ToLower(list[i]) < strings.ToLower(list[j]) 
  })
  for _, s := range list {
    names = append(names, makeSiteLink(s))
  }
  sl := List{names, port}
  err = app.execTempl(w, "allSites", sl)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func createSiteWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  if app.IsBusy() {
    <-app.web.trans
  }
  if r.Method == http.MethodGet {
    err := app.execTempl(w, "createSite", port)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
      http.Error(w, err.Error(), http.StatusBadRequest)
    }
    name := r.FormValue("name")
    login := r.FormValue("login")
    pass := r.FormValue("pass")
    link := r.FormValue("link")
    s := &Site{name, purecrypt.Symcode(login, app.web.word), purecrypt.Symcode(pass, app.web.word), link}
    err = app.AddSiteToDb(s)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    mp := fmt.Sprintf("Site %s was created", name)
    err = app.execTempl(w, "message", MessPort{mp, port})
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  } 
}

func showSiteWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    return
  }
  name := r.URL.Query().Get("name")
  defer r.Body.Close()
  s, err := app.GetSiteFromDb(name)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  dl := purecrypt.Desymcode(s.Login, app.web.word)
  dp := purecrypt.Desymcode(s.Pass, app.web.word)
  sw := Site{s.Name, dl, dp, s.Link}
  sp := SitePort{sw, port}
  err = app.execTempl(w, "oneSite", sp)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func deleteSiteWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    return
  }
  name := r.URL.Query().Get("name")
  defer r.Body.Close()
  err := app.RemoveSiteFromDb(strings.Trim(name, "\""))
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  ms := fmt.Sprintf("Site %s was deleted", name)
  err = app.execTempl(w, "message", MessPort{ms, port})
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func allCardsWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    return
  }
  list, err := app.GetAllCardsFromDb()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  names := []template.HTML{}
  sort.Slice(list, func(i, j int) bool { 
    return strings.ToLower(list[i].Name)[0] < strings.ToLower(list[j].Name)[0] 
  })
  for _, c := range list {
    names = append(names, makeCardLink(c))
  }
  listCards := List{names, port}
  err = app.execTempl(w, "allCards", listCards)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func createCardWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  if app.IsBusy() {
    <-app.web.trans
  }
  mistake := ""
  if r.Method == http.MethodGet {
    err := app.execTempl(w, "createCard", MistPort{"", port})
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
      http.Error(w, err.Error(), http.StatusBadRequest)
    }
    name := r.FormValue("name")
    number := r.FormValue("number")
    month := r.FormValue("month")
    year := r.FormValue("year")
    expire := fmt.Sprintf("%s / %s", month, year)
    cvc := r.FormValue("cvc")
    c := &Card{name, number, expire, cvc}
    mistake = c.CheckCard()
    if mistake != "" {
      mp := MistPort{mistake, port}
      err = app.execTempl(w, "createCard", mp)
      if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
      }
      return
    }
    encNum := purecrypt.Symcode(cleanNum(c.Number), app.web.word)
    encExp := purecrypt.Symcode(c.Expire, app.web.word)
    encCvc := purecrypt.Symcode(c.Cvc, app.web.word)
    ccr := Card{name, encNum, encExp, encCvc}
    err = app.AddCardToDb(ccr)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    m := fmt.Sprintf("Card %s was created", name)
    err = app.execTempl(w, "message", MessPort{m, port})
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  } 
}

func showCardWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    return
  }
  name := r.URL.Query().Get("name")
  c, err := app.GetCardFromDb(name)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  dn := purecrypt.Desymcode(c.Number, app.web.word)
  de := purecrypt.Desymcode(c.Expire, app.web.word)
  dc := purecrypt.Desymcode(c.Cvc, app.web.word)
  dcd := Card{c.Name, spaceNum(dn), de, dc}
  cp := CardPort{dcd, port}
  err = app.execTempl(w, "oneCard", cp)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func deleteCardWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    return
  }
  name := r.URL.Query().Get("name")
  err := app.RemoveCardFromDb(strings.Trim(name, "\""))
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  ms := fmt.Sprintf("Card %s was deleted", name)
  err = app.execTempl(w, "message", MessPort{ms, port})
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func createDocWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  if app.IsBusy() {
    <-app.web.trans
  }
  if r.Method == http.MethodGet {
    err := app.execTempl(w, "createDoc", port)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
    }
    nm := r.FormValue("name")
    val := r.FormValue("value")
    d := &Doc{nm, purecrypt.Symcode(val, app.web.word)}
    err = app.AddDocToDb(d)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    m := fmt.Sprintf("Doc %s was created", nm)
    err = app.execTempl(w, "message", MessPort{m, port})
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  }
}

func showDocsWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    return
  }
  dcs, err := app.GetDocsFromDb() 
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  ddx := []Doc{}
  for _, d := range dcs {
    dd := Doc{d.Name, purecrypt.Desymcode(d.Value, app.web.word)}
    ddx = append(ddx, dd)
  }
  dsp := DocsPort{ddx, port}
  err = app.execTempl(w, "allDocs", dsp)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}


func editDocWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    http.Redirect(w, r, "/", 302)
  }
  if app.IsBusy() {
    <-app.web.trans
  }
   doc := r.URL.Query().Get("doc")
   if r.Method == http.MethodGet {
     // ???
     dv := purecrypt.Desymcode(app.GetDocValue(doc), app.web.word)
      app.execTempl(w, "editDoc", DocPort{Doc{doc, dv}, port})
   }
   if r.Method == http.MethodPost {
     err := r.ParseForm()
     if err != nil {
       http.Error(w, err.Error(), http.StatusBadRequest)
     }
     del := r.FormValue("delete")
     if del == "del" {
       err = app.DeleteDocFromDb(doc)
       if err != nil {
         fmt.Println("DeleteDoc(): ", err)
         http.Error(w, err.Error(), http.StatusInternalServerError)
       }
       http.Redirect(w, r, Addr("docs"), 303)
     }
     val := r.FormValue("value")
     err = app.UpdateDocDb(doc, purecrypt.Symcode(val, app.web.word))
     if err != nil {
       fmt.Println("UpdateDoc(): ", err)
       http.Error(w, err.Error(), http.StatusInternalServerError)
     }
     http.Redirect(w, r, "/docs",303)
   }
}


func createPassrfWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  if app.IsBusy() {
    <-app.web.trans
  }
  if r.Method == http.MethodGet {
    err := app.execTempl(w, "createPassrf", port)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
    }
    sn := r.FormValue("serialnum")
    wn := r.FormValue("date")
    wm := r.FormValue("whom")
    cd := r.FormValue("code")
    wd := app.web.word
    ps := &PassRF{purecrypt.Symcode(sn, wd), purecrypt.Symcode(wn, wd),
      purecrypt.Symcode(wm, wd), purecrypt.Symcode(cd, wd)}
    err = app.AddPassrfToDb(ps)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    m := "PassRF was created"
    err = app.execTempl(w, "message", MessPort{m, port})
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  }
}

func showPassrfWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsDead() {
    return
  }
  p, err := app.GetPassrfFromDb()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  wd := app.web.word
  dp := PassRF{purecrypt.Desymcode(p.SerialNum, wd), purecrypt.Desymcode(p.Date, wd), 
    purecrypt.Desymcode(p.Whom, wd), purecrypt.Desymcode(p.Code, wd)}
  if dp.SerialNum == "" {
    http.Redirect(w, r, "/createPassrf", 302)
    return
  }
  pp := PassPort{dp, port} 
  err = app.execTempl(w, "passrf", pp)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

////   UTILS

func RecodeWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsBusy() {
    <-app.web.trans
  }
  if r.Method == http.MethodGet {
    if wc, ok := app.web.templs["wellcome"]; ok {
      wc.Execute(w, false)
    } 
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
      http.Error(w, err.Error(), http.StatusBadRequest)
    }
    word1 := r.FormValue("word1")
    word2 := r.FormValue("word2")
    ms := "Password length must be 5 or more"
    if len(word1) < 5 {
      mp := MessPort{ms, port}
      app.execTempl(w, "message", mp)
    }
    if word1 != word2 {
      m := "Passwords not matched"
      app.execTempl(w, "message", MessPort{m, port})
    }
    err = app.RecodeDb(word1)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    mg := "Passwords was changed"
    app.execTempl(w, "message", MessPort{mg, port})
  }
}

func BackupWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsBusy() {
    <-app.web.trans
  }
  err := app.BackupDb()
  if err != nil {
    terr := app.execTempl(w, "message", MessPort{fmt.Sprintf("%s", err), port})
    if terr != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
  }
  err = app.execTempl(w, "message", MessPort{"Done", port})
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func ShareWeb(w http.ResponseWriter, r *http.Request) {
  if app.IsBusy() {
    <-app.web.trans
  }
  err := DoJoinDb()
  if err != nil {
    app.execTempl(w, "message", MessPort{fmt.Sprintf("%s", err), port})
    return    
  }
  m := "Databases was synced"
  err = app.execTempl(w, "message", MessPort{m, port})
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func quitApp(w http.ResponseWriter, r *http.Request) {
  chanq <-struct{}{}
}