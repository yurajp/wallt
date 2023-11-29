package models

import (
  "strings"
  "net/http"
  "html/template"
  "database/sql"
  "context"
	"github.com/sirupsen/logrus"
)


type Web struct {
  Server *http.Server
  Ctx context.Context
  Templs map[string]*template.Template
  Word string
  Trans chan struct{}
  Log *logrus.Logger
  Quit chan struct{}
}

type Wdb struct {
  Db *sql.DB
}

type Site struct {
    Name string
    Login string
    Pass string
    Link string
}

type List struct {
  Names []template.HTML
}

type Export struct {
  Exists bool
  Port string
}

type Card struct {
    Name string
    Number string
    Expire string
    Cvc string
}

func (c *Card) CheckCard() string {
  if !checkNum(c.Number) {
    return "INCORRECT CARD NUMBER"
  }
  if !checkDate(strings.Split(c.Expire, " / ")) {
    return "INCORRECT EXPIRE DATE"
  }
  if !checkCvc(c.Cvc) {
    return "INCORRECT CVC"
  }
  return ""
}


type CardName struct {
  Name string
  Num string
}

type Doc struct {
  Name string
  Value string
}

type PassRF struct {
  SerialNum string
  Date string
  Whom string
  Code string
}

type Message struct {
  Text string
}
