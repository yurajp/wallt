package database

import (
    "fmt"
    "database/sql"
    
    "github.com/yurajp/wallt/internal/models"
)

type Wdb models.Wdb

var WDB *Wdb

func NewWdb(db *sql.DB) *Wdb {
  return &Wdb{db}
}

func (wdb *Wdb) CreateTables() error {
  query1 := `create table if not exists sites(name text primary key, login text, pass text, link text)`
  _, err := wdb.Db.Exec(query1)
  if err != nil {
    return err
  }
  query2 := `create table if not exists cards(name text primary key, number text, expire text, cvc text, unique(name, number))`
  _, err = wdb.Db.Exec(query2)
  if err != nil {
    return err
  }
  query3 := `create table if not exists docs(name text primary key, value text)`
  _, err = wdb.Db.Exec(query3)
  if err != nil {
    return err
  }
  query4 := `create table if not exists passrf(serialnum text primary key, date text, whom text, code text)`
  _, err = wdb.Db.Exec(query4)
  if err != nil {
    return err
  }
  return nil
}

func (wdb *Wdb) AddSiteToDb(s *models.Site) error {
    query := `insert into sites(name, login, pass, link) values(?, ?, ?, ?) on conflict(name) do update set pass=excluded.pass, login=excluded.login, link=excluded.link`
    _, err := wdb.Db.Exec(query, s.Name, s.Login, s.Pass, s.Link)
    if err != nil {
      return err
    }
    return nil
}

func (wdb *Wdb) GetSiteFromDb(q string) (models.Site, error) {
  query := `select name, login, pass, link from sites where lower(name) like ?`
  row := wdb.Db.QueryRow(query, q + "%")
  var s models.Site
  err := row.Scan(&s.Name, &s.Login, &s.Pass, &s.Link)
  if err != nil {
    return models.Site{}, err
  }
  return s, nil
}


func (wdb *Wdb) RemoveSiteFromDb(s string) error {
  query := `delete from sites where name=?`
  _, err := wdb.Db.Exec(query, s)
  if err != nil {
    return err
  }
  return nil
}

func (wdb *Wdb) GetAllSitesFromDb() ([]string, error) {
  query := `select name from sites`
  rows, err := wdb.Db.Query(query)
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

func (wdb *Wdb) GetAllCardsFromDb() ([]models.CardName, error) {
  query := `select name, number from cards`
  rows, err := wdb.Db.Query(query)
  if err != nil {
    return []models.CardName{}, err
  }
  defer rows.Close()
  list := []models.CardName{}
  for rows.Next() {
    var cn models.CardName
    rows.Scan(&cn.Name, &cn.Num)
    list = append(list, cn)
  }
  return list, nil
}

func (wdb *Wdb) AddCardToDb(c models.Card) error {
    query := `insert into cards(name, number, expire, cvc) values(?, ?, ?, ?) on conflict(name) do update set number=excluded.number, expire=excluded.expire, cvc=excluded.cvc`
     _, err := wdb.Db.Exec(query, c.Name, c.Number, c.Expire, c.Cvc)
    if err != nil {
      return err
    }
    return nil
}

func (wdb *Wdb) GetCardFromDb(q string) (models.Card, error) {
  query := `select name, number, expire, cvc from cards where name=?`
  row := wdb.Db.QueryRow(query, q)
  var c models.Card
  err := row.Scan(&c.Name, &c.Number, &c.Expire, &c.Cvc)
  if err != nil {
    return models.Card{}, err
  }
  return c, nil
}

func (wdb *Wdb) RemoveCardFromDb(c string) error {
  query := `delete from cards where name=?`
  _, err := wdb.Db.Exec(query, c)
  if err != nil {
    return err
  }
  return nil
}

func (wdb *Wdb) AddDocToDb(d *models.Doc) error {
  query := `insert into docs(name, value) values(?, ?)`
  _, err := wdb.Db.Exec(query, d.Name, d.Value)
  if err != nil {
    return err
  }
  return nil
}

func (wdb *Wdb) AddPassrfToDb(p *models.PassRF) error {
  query := `insert into passrf(serialnum, date, whom, code) values(?, ?, ?, ?)`
  sn := p.SerialNum
  wn := p.Date
  wm := p.Whom
  cd := p.Code
  _, err := wdb.Db.Exec(query, sn, wn, wm, cd)
  if err != nil {
    return err
  }
  return nil
}

func (wdb *Wdb) GetDocsFromDb() ([]models.Doc, error) {
  query := `select * from docs`
  rows, err := wdb.Db.Query(query)
  if err != nil {
    fmt.Println(err)
    return []models.Doc{}, err
  }
  defer rows.Close()
  dcs := []models.Doc{}
  for rows.Next() {
    var d models.Doc
    rows.Scan(&d.Name, &d.Value)
    dcs = append(dcs, d)
  }
  return dcs, nil
}

func (wdb *Wdb) GetPassrfFromDb() (models.PassRF, error) {
  query := `select * from passrf`
  rows, err := wdb.Db.Query(query)
  if err != nil {
    return models.PassRF{}, err
  }
  defer rows.Close()
  var p models.PassRF
  for rows.Next() {
    var tp models.PassRF
    rows.Scan(&tp.SerialNum, &tp.Date, &tp.Whom, &tp.Code)
    p = tp
  }
  return p, nil
}

func (wdb *Wdb) GetDocValue(d string) string {
  query := `select name, value from docs where name = ?`
  row := wdb.Db.QueryRow(query, d)
  var doc models.Doc
  row.Scan(&doc.Name, &doc.Value)
  return doc.Value
}

func (wdb *Wdb) DeleteDocFromDb(d string) error {
  query := `delete from docs where name = ?`
  _, err := wdb.Db.Exec(query, d)
  if err != nil {
    return err
  }
  return nil
}

func (wdb *Wdb) UpdateDocDb(name, val string) error {
  
  sttm := `update docs set value = ? where name = ?`
  _, err := wdb.Db.Exec(sttm, name, val)
  if err != nil {
    return err
  }
  return nil
}
