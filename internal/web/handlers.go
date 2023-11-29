package web

import (
    "fmt"
    "net/http"
    "sort"
    "strings"
    "html/template"
  
    "github.com/yurajp/wallt/internal/models"
    "github.com/yurajp/wallt/internal/purecrypt"
    "github.com/yurajp/wallt/internal/database"
)

func wellcome(w http.ResponseWriter, r *http.Request) {
  exists := purecrypt.ChWordExists()
  if r.Method == http.MethodGet {
    if wc, ok := WEB.Templs["wellcome"]; ok {
      exp := models.Export{exists, WEB.Server.Addr}
      
      wc.Execute(w, exp)
    }
  }
  if r.Method == http.MethodPost {
    if !exists {
      err := database.WDB.CreateTables()
      if err != nil {
      	WEB.Log.WithError(err).Error("Cannot create tables")
      }
      err = r.ParseForm()
      if err != nil {
      	WEB.Log.WithError(err).Error("Cannot parse form")
      }
      word1 := r.FormValue("word1")
      word2 := r.FormValue("word2")
      if len(word1) < 5 || word1 != word2 {
        http.Redirect(w, r, "/", 302)
        return
      }
      err = purecrypt.WriteCheckword(word1)
      if err != nil {
      	WEB.Log.WithError(err).Error("Cannot write checkword")
      }
      WEB.Word = word1
      http.Redirect(w, r, "/home", 302)
    } else {
      err := r.ParseForm()
      if err != nil {
      	WEB.Log.WithError(err).Error("Cannot parse form")
      }
      word := r.FormValue("word")
      if purecrypt.IsCorrect(word) {
        WEB.Word = word
        http.Redirect(w, r, "/home", 303)
      } else {
        WEB.Log.Warn("Wrong Password")
        http.Redirect(w, r, "/", 302)
      }
    }
  }
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
  if WEB.IsDead() {
    return
  }
  w.Header().Set("Cache-Control", "no-store")
  
  addr := WEB.Server.Addr
  err := WEB.ExecTempl(w, "home", addr)
  if err != nil {
  	WEB.Log.WithError(err).Error("Cannot execute home template")
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func exitWeb(w http.ResponseWriter, r *http.Request) {
  WEB.Dies()
  http.Redirect(w, r, "/", 302)
  if len(WEB.Trans) == 1 {
    <-WEB.Trans
  }
}

func allSitesWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsDead() {
    WEB.Log.Warn("APP is DEAD!")
    return
  }
  list, err := database.WDB.GetAllSitesFromDb()
  if err != nil {
  	WEB.Log.WithError(err).Error("Cannot get sites from DB")
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  names := []template.HTML{}
  sort.Slice(list, func(i, j int) bool { 
    return strings.ToLower(list[i]) < strings.ToLower(list[j]) 
  })
  for _, s := range list {
    names = append(names, WEB.MakeSiteLink(s))
  }
  sl := List{&models.List{names}}
  err = WEB.ExecTempl(w, "allSites", sl)
  if err != nil {
  	WEB.Log.WithError(err).Error("Cannot execute sites template")
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func createSiteWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  if WEB.IsBusy() {
    <-WEB.Trans
  }
  addr := WEB.Server.Addr
  if r.Method == http.MethodGet {
    err := WEB.ExecTempl(w, "createSite", addr)
    if err != nil {
    	WEB.Log.WithError(err).Error("Cannot execute create site template")
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
    	WEB.Log.WithError(err).Error("Cannot parse form")
      http.Error(w, err.Error(), http.StatusBadRequest)
    }
    name := r.FormValue("name")
    login := r.FormValue("login")
    pass := r.FormValue("pass")
    link := r.FormValue("link")
    s := &models.Site{name, purecrypt.Symcode(login, WEB.Word), purecrypt.Symcode(pass, WEB.Word), link}
    err = database.WDB.AddSiteToDb(s)
    if err != nil {
    	WEB.Log.WithError(err).Error("Cannot add site to DB")
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    mp := fmt.Sprintf("Site %s was created", name)
    err = WEB.ExecTempl(w, "message", Message{&models.Message{mp}})
    if err != nil {
    	WEB.Log.WithError(err).Warn("Cannot execute message template")
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  } 
}

func showSiteWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsDead() {
    return
  }
  name := r.URL.Query().Get("name")
  defer r.Body.Close()
  s, err := database.WDB.GetSiteFromDb(name)
  if err != nil {
  	WEB.Log.WithError(err).Error("Cannot create site in DB")
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  dl := purecrypt.Desymcode(s.Login, WEB.Word)
  dp := purecrypt.Desymcode(s.Pass, WEB.Word)
  sw := Site{&models.Site{s.Name, dl, dp, s.Link}}
  err = WEB.ExecTempl(w, "oneSite", sw)
  if err != nil {
  	WEB.Log.WithError(err).Error("Cannot execute site template")
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func deleteSiteWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsDead() {
    return
  }
  name := r.URL.Query().Get("name")
  defer r.Body.Close()
  err := database.WDB.RemoveSiteFromDb(strings.Trim(name, "\""))
  if err != nil {
  	WEB.Log.WithError(err).Error("Cannot remove site from DB")
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  ms := fmt.Sprintf("Site %s was deleted", name)
  err = WEB.ExecTempl(w, "message", Message{&models.Message{ms}})
  if err != nil {
  	WEB.Log.WithError(err).Warn("Cannot execute message template")
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func allCardsWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsDead() {
    return
  }
  list, err := database.WDB.GetAllCardsFromDb()
  if err != nil {
  WEB.Log.WithError(err).Error("Cannot get cards from DB")
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  names := []template.HTML{}
  sort.Slice(list, func(i, j int) bool { 
    return strings.ToLower(list[i].Name)[0] < strings.ToLower(list[j].Name)[0] 
  })
  for _, c := range list {
    names = append(names, WEB.MakeCardLink(c))
  }
  listCards := List{&models.List{names}}
  err = WEB.ExecTempl(w,
  "allCards", listCards)
  if err != nil {
  WEB.Log.WithError(err).Error("Cannot execute cards template")
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func createCardWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  if WEB.IsBusy() {
    <-WEB.Trans
  }
  mistake := ""
  if r.Method == http.MethodGet {
    err := WEB.ExecTempl(w, "createCard", Message{&models.Message{""}})
    if err != nil {
  WEB.Log.WithError(err).Error("Cannot execute create card template")
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
      WEB.Log.WithError(err).Error("Cannot parse form")
      http.Error(w, err.Error(), http.StatusBadRequest)
    }
    name := r.FormValue("name")
    number := r.FormValue("number")
    month := r.FormValue("month")
    year := r.FormValue("year")
    expire := fmt.Sprintf("%s / %s", month, year)
    cvc := r.FormValue("cvc")
    c := Card{&models.Card{name, number, expire, cvc}}
    mistake = c.CheckCard()
    if mistake != "" {
      ms := Message{&models.Message{mistake}}
      err = WEB.ExecTempl(w, "createCard", ms)
      if err != nil {
        WEB.Log.WithError(err).Error("Cannot execute create card template")
        http.Error(w, err.Error(), http.StatusInternalServerError)
      }
      return
    }
    encNum := purecrypt.Symcode(cleanNum(c.Number), WEB.Word)
    encExp := purecrypt.Symcode(c.Expire, WEB.Word)
    encCvc := purecrypt.Symcode(c.Cvc, WEB.Word)
    ccr := models.Card{name, encNum, encExp, encCvc}
    err = database.WDB.AddCardToDb(ccr)
    if err != nil {
      WEB.Log.WithError(err).Error("Cannot add card to DB")
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    m := fmt.Sprintf("Card %s was created", name)
    err = WEB.ExecTempl(w, "message", Message{&models.Message{m}})
    if err != nil {
       WEB.Log.WithError(err).Warn("Cannot execute message template")
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  } 
}

func showCardWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsDead() {
    return
  }
  name := r.URL.Query().Get("name")
  c, err := database.WDB.GetCardFromDb(name)
  if err != nil {
    WEB.Log.WithError(err).Error("Cannot get card from DB")
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  dn := purecrypt.Desymcode(c.Number, WEB.Word)
  de := purecrypt.Desymcode(c.Expire, WEB.Word)
  dc := purecrypt.Desymcode(c.Cvc, WEB.Word)
  dcd := Card{&models.Card{c.Name, spaceNum(dn), de, dc}}
  err = WEB.ExecTempl(w, "oneCard", dcd)
  if err != nil {
    WEB.Log.WithError(err).Error("Cannot execute card template")
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func deleteCardWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsDead() {
    return
  }
  name := r.URL.Query().Get("name")
  err := database.WDB.RemoveCardFromDb(strings.Trim(name, "\""))
  if err != nil {
    WEB.Log.WithError(err).Error("Cannot remove from DB")
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  ms := fmt.Sprintf("Card %s was deleted", name)
  err = WEB.ExecTempl(w, "message", Message{&models.Message{ms}})
  if err != nil {
    WEB.Log.WithError(err).Warn("Cannot execute message template")
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func createDocWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  if WEB.IsBusy() {
    <-WEB.Trans
  }
  addr := WEB.Server.Addr
  if r.Method == http.MethodGet {
    err := WEB.ExecTempl(w, "createDoc", addr)
    if err != nil {
      WEB.Log.WithError(err).Error("Cannot execute create doc template")
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
      WEB.Log.WithError(err).Error("Cannot parse form")
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
    }
    nm := r.FormValue("name")
    val := r.FormValue("value")
    d := &models.Doc{nm, purecrypt.Symcode(val, WEB.Word)}
    err = database.WDB.AddDocToDb(d)
    if err != nil {
      WEB.Log.WithError(err).Error("Cannot add doc to DB")
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    m := fmt.Sprintf("Doc %s was created", nm)
    err = WEB.ExecTempl(w, "message", Message{&models.Message{m}})
    if err != nil {
      WEB.Log.WithError(err).Warn("Cannot execute message template")
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  }
}

func showDocsWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsDead() {
    return
  }
  dcs, err := database.WDB.GetDocsFromDb() 
  if err != nil {
    WEB.Log.WithError(err).Error("Cannot get docs from DB")
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  ddx := []*models.Doc{}
  for _, d := range dcs {
    dd := models.Doc{d.Name, purecrypt.Desymcode(d.Value, WEB.Word)}
    ddx = append(ddx, &dd)
  }
  dsp := Docs{ddx}
  err = WEB.ExecTempl(w, "allDocs", dsp)
  if err != nil {
    WEB.Log.WithError(err).Error("Cannot execute docs template")
 //   http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}


func editDocWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsDead() {
    http.Redirect(w, r, "/", 302)
  }
  if WEB.IsBusy() {
    <-WEB.Trans
  }
   doc := r.URL.Query().Get("doc")
   if r.Method == http.MethodGet {

     dv := purecrypt.Desymcode(database.WDB.GetDocValue(doc), WEB.Word)
      WEB.ExecTempl(w, "editDoc", Doc{&models.Doc{doc, dv}})
   }
   if r.Method == http.MethodPost {
     err := r.ParseForm()
     if err != nil {
       WEB.Log.WithError(err).Error("Cannot parse form")
       http.Error(w, err.Error(), http.StatusBadRequest)
     }
     del := r.FormValue("delete")
     if del == "del" {
       err = database.WDB.DeleteDocFromDb(doc)
       if err != nil {
         WEB.Log.WithError(err).Error("Cannot delete doc from DB")
         http.Error(w, err.Error(), http.StatusInternalServerError)
       }
       http.Redirect(w, r, "/docs", 303)
     }
     val := r.FormValue("value")
     err = database.WDB.UpdateDocDb(doc, purecrypt.Symcode(val, WEB.Word))
     if err != nil {
       WEB.Log.WithError(err).Error("Cannot update doc in DB")
       http.Error(w, err.Error(), http.StatusInternalServerError)
     }
     http.Redirect(w, r, "/docs",303)
   }
}


func createPassrfWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsDead() {
    http.Redirect(w, r, "/", 302)
    return
  }
  if WEB.IsBusy() {
    <-WEB.Trans
  }
  addr := WEB.Server.Addr
  if r.Method == http.MethodGet {
    err := WEB.ExecTempl(w, "createPassrf", addr)
    if err != nil {
      WEB.Log.WithError(err).Error("Cannot execute create PF template")
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
      WEB.Log.WithError(err).Error("Cannot parse form")
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
    }
    sn := r.FormValue("serialnum")
    wn := r.FormValue("date")
    wm := r.FormValue("whom")
    cd := r.FormValue("code")
    wd := WEB.Word
    ps := &models.PassRF{purecrypt.Symcode(sn, wd), purecrypt.Symcode(wn, wd),
      purecrypt.Symcode(wm, wd), purecrypt.Symcode(cd, wd)}
    err = database.WDB.AddPassrfToDb(ps)
    if err != nil {
      WEB.Log.WithError(err).Error("Cannot add passport to DB")
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    m := "PassRF was created"
    err = WEB.ExecTempl(w, "message", Message{&models.Message{m}})
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  }
}

func showPassrfWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsDead() {
    return
  }
  p, err := database.WDB.GetPassrfFromDb()
  if err != nil {
    WEB.Log.WithError(err).Error("Cannot get passport from DB")
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  wd := WEB.Word
  dp := models.PassRF{purecrypt.Desymcode(p.SerialNum, wd), purecrypt.Desymcode(p.Date, wd), 
    purecrypt.Desymcode(p.Whom, wd), purecrypt.Desymcode(p.Code, wd)}
  if dp.SerialNum == "" {
    http.Redirect(w, r, "/createPassrf", 302)
    return
  }
  err = WEB.ExecTempl(w, "passrf", PassRF{&dp})
  if err != nil {
    WEB.Log.WithError(err).Error("Cannot execute passport template")
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

////   UTILS

func RecodeWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsBusy() {
    <-WEB.Trans
  }
  if r.Method == http.MethodGet {
    if wc, ok := WEB.Templs["wellcome"]; ok {
      wc.Execute(w, false)
    } 
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
      WEB.Log.WithError(err).Error("Cannot parse form")
      http.Error(w, err.Error(), http.StatusBadRequest)
    }
    word1 := r.FormValue("word1")
    word2 := r.FormValue("word2")
    ms := "Password length must be 5 or more"
    if len(word1) < 5 {
      mp := Message{&models.Message{ms}}
      WEB.ExecTempl(w, "message", mp)
    }
    if word1 != word2 {
      m := "Passwords not matched"
      WEB.ExecTempl(w, "message", Message{&models.Message{m}})
    }
    err = database.WDB.RecodeDb(WEB.Word, word1)
    if err != nil {
      WEB.Log.WithError(err).Error("Cannot recode DB")
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    
    mg := "Passwords was changed. Please restart Wallt"
    WEB.ExecTempl(w, "message", Message{&models.Message{mg}})
  }
}

func BackupWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsBusy() {
    <-WEB.Trans
  }
//  addr := WEB.Server.Addr
  err := database.BackupDb()
  if err != nil {
    WEB.Log.WithError(err).Error("Cannot backup DB")
    terr := WEB.ExecTempl(w, "message", Message{&models.Message{fmt.Sprintf("%s", err)}})
    if terr != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
  }
  err = WEB.ExecTempl(w, "message", Message{&models.Message{"Done"}})
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func ShareWeb(w http.ResponseWriter, r *http.Request) {
  if WEB.IsBusy() {
    <-WEB.Trans
  }
  err := database.DoJoinDb()
  if err != nil {
    WEB.Log.WithError(err).Error("Cannot join DBs")
    WEB.ExecTempl(w, "message", Message{&models.Message{fmt.Sprintf("%s", err)}})
    return    
  }
  m := "Databases was synced"
  err = WEB.ExecTempl(w, "message", Message{&models.Message{m}})
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func quitApp(w http.ResponseWriter, r *http.Request) {
  WEB.Quit <-struct{}{}
}