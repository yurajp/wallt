package main

import (
  "fmt"
  "net/http"
  "errors"
  "strings"
  "html/template"
  "regexp"
  "strconv"
  
  "github.com/yurajp/wallt/purecrypt"
)


func check(err error) {
  if err != nil {
    panic(err)
  }
}

func Addr(s string) string {
  return fmt.Sprintf("https://localhost%s/%s", app.web.server.Addr, s)
}

func (app *App) IsDead() bool {
  return app.web.word == ""
}

func (app *App) Dies() {
  app.web.word = ""
}

func (app *App) IsBusy() bool {
  return len(app.web.trans) == 1
}

func (app *App) execTempl(w http.ResponseWriter, t string, data any) error {
  if app.IsDead() {
    return errors.New("Password not set")
  }
  if tmp, ok := app.web.templs[t]; ok {
    err := tmp.Execute(w, data)
    if err != nil {
      return err
    }
  } else {
    return errors.New("Template does not exists")
  }
  return nil
}

func makeSiteLink(nm string) template.HTML {
  addr := app.web.server.Addr
  url := fmt.Sprintf(`<a href='#' onclick="openRow('http://localhost%s/site?name=%s')">%s</a>`, addr, nm, nm)
  return template.HTML(url)
}

func cleanNum(n string) string {
  return strings.Replace(n, " ", "", -1)
}
  
func makeCardLink(cn CardName) template.HTML {
  if cn.Num == "" {
    return template.HTML("")
  }
  url := fmt.Sprintf(`<a href="http://localhost:8686/card?name=%s" onclick="openRow(self.href)" target="_self">%s</a>`, cn.Name, cn.Name)
  dcN := purecrypt.Desymcode(cn.Num, app.web.word)
  shN := "*" + dcN[12:]
  span := fmt.Sprintf(`<span>%s</span>`, shN)
  return template.HTML(url + span)
}

func spaceNum(n string) string {
  s := " "
  return n[:4] + s + n[4:8] + s + n[8:12] + s + n[12:]
}

func makeExpire(n string) (string, bool) {
  re := regexp.MustCompile(`\d\d[/ -\.]\d\d`)
  if !re.MatchString(n) {
    return "", false
  }
  sp := regexp.MustCompile(`[\./ -]`)
  exs := sp.Split(n, -1)
  return fmt.Sprintf("%s / %s", exs[0], exs[1]), true
}

func checkNum(n string) bool {
  re := regexp.MustCompile(`\d{4}\s?\d{4}\s?\d{4}\s?\d{4}`)
  return re.MatchString(n)
}  

func checkCvc(n string) bool {
  re := regexp.MustCompile(`\d\d\d`)
  return re.MatchString(n)
}

func checkDate(my []string) bool {
  if len(my) != 2 {
    return false
  }
  re := regexp.MustCompile(`\d\d`)
  if !re.MatchString(my[0]) || !re.MatchString(my[1]) {
    return false
  }
  dm, _ := strconv.Atoi(my[0])
  dy, _ := strconv.Atoi(my[1])
  if dm < 1 || dm > 12 || dy < 23 || dy > 35 {
    return false
  }
  return true
}
