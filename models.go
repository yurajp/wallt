package main

import (
  "strings"
  "net/http"
  "html/template"
  "database/sql"
  "context"
)

type App struct {
  web *Web
  db *sql.DB
}

type Web struct {
  server *http.Server
  ctx context.Context
  templs map[string]*template.Template
  word string
  trans chan struct{}
}

type Site struct {
    Name string
    Login string
    Pass string
    Link string
}

type List struct {
  Names []template.HTML
  Port string
}

type SitePort struct {
  Site
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

type CardPort struct {
  Card
  Port string
}

type CardName struct {
  Name string
  Num string
}

type Doc struct {
  Name string
  Value string
}

type DocsPort struct {
  Docs []Doc
  Port string
}

type PassRF struct {
  SerialNum string
  Date string
  Whom string
  Code string
}

type PassPort struct {
  PassRF
  Port string
}

type MistPort struct {
  Message string
  Port string
}

type MessPort struct {
  Message string
  Port string
}
