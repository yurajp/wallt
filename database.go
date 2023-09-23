package main

import (
    "fmt"
    "database/sql"
    "time"
    "errors"
    "os"
    "io"
    "os/exec"
  
    _ "github.com/mattn/go-sqlite3"
    "github.com/yurajp/wallt/purecrypt"
)


func (app *App) createTables() error {
  query1 := `create table if not exists sites(name text primary key, login text, pass text, link text)`
  _, err := app.db.Exec(query1)
  if err != nil {
    return err
  }
  query2 := `create table if not exists cards(name text primary key, number text, expire text, cvc text, unique(name, number))`
  _, err = app.db.Exec(query2)
  if err != nil {
    return err
  }
  query3 := `create table if not exists docs(name text primary key, value text)`
  _, err = app.db.Exec(query3)
  if err != nil {
    return err
  }
  query4 := `create table if not exists passrf(serialnum text primary key, date text, whom text, code text)`
  _, err = app.db.Exec(query4)
  if err != nil {
    return err
  }
  return nil
}

func (app *App) AddSiteToDb(s *Site) error {
    query := `insert into sites(name, login, pass, link) values(?, ?, ?, ?) on conflict(name) do update set pass=excluded.pass, login=excluded.login, link=excluded.link`
    _, err := app.db.Exec(query, s.Name, s.Login, s.Pass, s.Link)
    if err != nil {
      return err
    }
    return nil
}

func (app *App) GetSiteFromDb(q string) (Site, error) {
  query := `select name, login, pass, link from sites where lower(name) like ?`
  row := app.db.QueryRow(query, q + "%")
  var s Site
  err := row.Scan(&s.Name, &s.Login, &s.Pass, &s.Link)
  if err != nil {
    return Site{}, err
  }
  return s, nil
}


func (app *App) RemoveSiteFromDb(s string) error {
  query := `delete from sites where name=?`
  _, err := app.db.Exec(query, s)
  if err != nil {
    return err
  }
  return nil
}

func (app *App) GetAllSitesFromDb() ([]string, error) {
  query := `select name from sites`
  rows, err := app.db.Query(query)
  if err != nil {
    return []string{}, err
  }
  defer rows.Close()
  list := []string{}
  for rows.Next() {
    var s string
    rows.Scan(&s)
    list = append(list, s)
  }
  return list, nil
}

func (app *App) GetAllCardsFromDb() ([]CardName, error) {
  query := `select name, number from cards`
  rows, err := app.db.Query(query)
  if err != nil {
    return []CardName{}, err
  }
  defer rows.Close()
  list := []CardName{}
  for rows.Next() {
    var cn CardName
    rows.Scan(&cn.Name, &cn.Num)
    list = append(list, cn)
  }
  return list, nil
}

func (app *App) AddCardToDb(c Card) error {
    query := `insert into cards(name, number, expire, cvc) values(?, ?, ?, ?) on conflict(name) do update set number=excluded.number, expire=excluded.expire, cvc=excluded.cvc`
     _, err := app.db.Exec(query, c.Name, c.Number, c.Expire, c.Cvc)
    if err != nil {
      return err
    }
    return nil
}

func (app *App) GetCardFromDb(q string) (Card, error) {
  query := `select name, number, expire, cvc from cards where name=?`
  row := app.db.QueryRow(query, q)
  var c Card
  err := row.Scan(&c.Name, &c.Number, &c.Expire, &c.Cvc)
  if err != nil {
    return Card{}, err
  }
  return c, nil
}

func (app *App) RemoveCardFromDb(c string) error {
  query := `delete from cards where name=?`
  _, err := app.db.Exec(query, c)
  if err != nil {
    return err
  }
  return nil
}

func (app *App) AddDocToDb(d *Doc) error {
  query := `insert into docs(name, value) values(?, ?)`
  _, err := app.db.Exec(query, d.Name, d.Value)
  if err != nil {
    return err
  }
  return nil
}

func (app *App) AddPassrfToDb(p *PassRF) error {
  query := `insert into passrf(serialnum, date, whom, code) values(?, ?, ?, ?)`
  sn := p.SerialNum
  wn := p.Date
  wm := p.Whom
  cd := p.Code
  _, err := app.db.Exec(query, sn, wn, wm, cd)
  if err != nil {
    return err
  }
  return nil
}

func (app *App) GetDocsFromDb() ([]Doc, error) {
  query := `select * from docs`
  rows, err := app.db.Query(query)
  if err != nil {
    fmt.Println(err)
    return []Doc{}, err
  }
  defer rows.Close()
  dcs := []Doc{}
  for rows.Next() {
    var d Doc
    rows.Scan(&d.Name, &d.Value)
    dcs = append(dcs, d)
  }
  return dcs, nil
}

func (app *App) GetPassrfFromDb() (PassRF, error) {
  query := `select * from passrf`
  rows, err := app.db.Query(query)
  if err != nil {
    return PassRF{}, err
  }
  defer rows.Close()
  var p PassRF
  for rows.Next() {
    var tp PassRF
    rows.Scan(&tp.SerialNum, &tp.Date, &tp.Whom, &tp.Code)
    p = tp
  }
  return p, nil
}

func (app *App) RecodeDb(newWord string) error {
    nDb, err := sql.Open("sqlite3", "temp.db")
    if err != nil {
       return err
    }
    defer nDb.Close()
    nWb := &Web{}
    nAp := &App{nWb, nDb}
    nAp.createTables()
    
    querySs := `select * from sites`
    queryA := `insert in sites(name, login, pass, link) values(?, ?, ?, ?)`
    rowsSs, err := app.db.Query(querySs)
    if err != nil {
      return err
    }
    for rowsSs.Next() {
      var nm, lg, ps, lk string
      rowsSs.Scan(&nm, &lg, &ps, &lk)
      nlg := purecrypt.Symcode(purecrypt.Desymcode(lg, app.web.word), newWord)
      nps := purecrypt.Symcode(purecrypt.Desymcode(ps, app.web.word), newWord)
      _, err = nDb.Exec(queryA, nm, nlg, nps, lk)
      if err != nil {
        return err
      }
    }
    queryCs := `select * from cards`
    queryB := `insert in cards(name, number, expire, cvc) values(?, ?, ?, ?)`
    rowsCs, err := app.db.Query(queryCs)
    if err != nil {
      return err
    }
    for rowsCs.Next() {
      var nc, nb, ex, cv string
      rowsCs.Scan(&nc, &nb, &ex, &cv)
      nnb := purecrypt.Symcode(purecrypt.Desymcode(nb, app.web.word), newWord)
      nex := purecrypt.Symcode(purecrypt.Desymcode(ex, app.web.word), newWord)
      ncv := purecrypt.Symcode(purecrypt.Desymcode(cv, app.web.word), newWord)
      _, err = nDb.Exec(queryB, nc, nnb, nex, ncv)
      if err != nil {
        return err
      }
    }
    nDb.Close()
    app.db.Close()
    app.db = nil
    err = os.Rename("temp.db", "wallx.db")
    if err != nil {
      return err
    }
    db, err := sql.Open("sqlite3", "wallx.db")
    if err != nil {
      return err
    }
    app.db = db
    app.web.word = newWord
    err = purecrypt.WriteCheckword(newWord)
    if err != nil {
      return err
    }
    return nil
}

func (app *App) BackupDb() error {
  ty, tm, td := time.Now().Date()
  date := fmt.Sprintf("%v%v%v", ty - 2000, tm, td)
  i, err := os.Stat("archive")
  if errors.Is(err, os.ErrNotExist) || !i.IsDir() {
    err = os.Mkdir("archive", 0775)
    if err != nil {
      return fmt.Errorf(" Cannot make dir:\n%s", err)
    }
  }
  fpath := fmt.Sprintf("archive/%s.wallt.db", date)
  f, err := os.Create(fpath)
  if err != nil {
    return err
  }
  defer f.Close()
  fdb, err := os.Open("wallt.db")
  if err != nil {
    return err
  }
  defer fdb.Close()
  _, err = io.Copy(f, fdb)
  if err != nil {
    return err
  }
  return nil
}

func (app *App) ShareDb() error {
  wd, _ := os.Getwd()
  upld := "/storage/emulated/0/Uploads/"
  cmd := exec.Command("cp", "wallt.db", upld)
  err := cmd.Run()
  if err != nil {
    return fmt.Errorf(" Cannot copy\n %s",err)
  }
  err = os.Chdir("../../messer-mobile")
  if err != nil {
    return fmt.Errorf(" Cannot cd to messer\n %s", err)
  }
  
  cmd = exec.Command("./messer-mobile", "up")
  err = cmd.Run()
  if err != nil {
    return fmt.Errorf(" Cannot start uploading\n %s", err)
  }
  err = os.Chdir(wd)
  if err != nil {
    return fmt.Errorf(" Cannot return to Wallt\n %s", err)
  }
  return nil
}
