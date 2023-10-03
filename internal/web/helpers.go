package web

import (
  "fmt"
  "net/http"
  "errors"
  "strings"
  "html/template"
  "regexp"
  
  "github.com/yurajp/wallt/internal/purecrypt"
  "github.com/yurajp/wallt/internal/models"
)


func check(err error) {
  if err != nil {
    panic(err)
  }
}

func (web *Web) Addr(s string) string {
  return fmt.Sprintf("https://localhost%s/%s", web.Server.Addr, s)
}

func (web *Web) IsDead() bool {
  return web.Word == ""
}

func (web *Web) Dies() {
  web.Word = ""
}

func (web *Web) IsBusy() bool {
  return len(web.Trans) == 1
}

func (web *Web) ExecTempl(w http.ResponseWriter, t string, data any) error {
  if web.IsDead() {
    return errors.New("Password not set")
  }
  if tmp, ok := web.Templs[t]; ok {
    err := tmp.Execute(w, data)
    if err != nil {
      return err
    }
  } else {
    return errors.New("Template does not exists")
  }
  return nil
}

func (web *Web) MakeSiteLink(nm string) template.HTML {
  addr := web.Server.Addr
  url := fmt.Sprintf(`<a href='#' onclick="openRow('http://localhost%s/site?name=%s')">%s</a>`, addr, nm, nm)
  return template.HTML(url)
}

func cleanNum(n string) string {
  return strings.Replace(n, " ", "", -1)
}
  
func (web *Web) MakeCardLink(cn models.CardName) template.HTML {
  if cn.Num == "" {
    return template.HTML("")
  }
  addr := web.Server.Addr
  url := fmt.Sprintf(`<a href="http://localhost%s/card?name=%s" onclick="openRow(self.href)" target="_self">%s</a>`, addr, cn.Name, strings.ToUpper(cn.Name))
  dcN := purecrypt.Desymcode(cn.Num, web.Word)
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
