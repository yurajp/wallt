package main

import (
  "os"
  "fmt"
  "database/sql"
 // "github.com/yurajp/wallt/config"
  "github.com/melbahja/goph"
  _ "github.com/mattn/go-sqlite3"
)

var (
  tempath = "remote.db"
  lockdb = "wallt.db"
  keypath = "/data/data/com.termux/files/home/.ssh/id_rsa"
  user = "yura"
  raddr = "192.168.1.38"
  rdbpath = "/home/yura/golangs/wallt/wallt.db"
)

func getRemoteDb() error {
  if _, err := os.Stat(tempath); err != nil {
    if os.IsNotExist(err) {
      f, er := os.Create(tempath)
      if er != nil {
        return er
      }
      defer f.Close()
    }
  }
  auth, err := goph.Key(keypath, "")
  if err != nil {
    return fmt.Errorf("key: %s", err)
  }
	client, err := goph.New(user, raddr, auth)
	if err != nil {
		return fmt.Errorf("ssh: %s", err)
	}
	defer client.Close()
  
  err = client.Download(rdbpath , tempath)
  if err != nil {
    return fmt.Errorf("download db: %s", err)
  }
  return nil
}

func loadSites(dbpath string) ([]Site, error) {
  sites := []Site{}
  db, err := sql.Open("sqlite3", dbpath)
  if err != nil {
    return sites, fmt.Errorf("open %s db: %s", dbpath, err)
  }
  defer db.Close()
  query := "SELECT * FROM sites"
  rows, err := db.Query(query)
  if err != nil {
    return sites, fmt.Errorf("db query: %s", err)
  }
  defer rows.Close()
  for rows.Next() {
    var st Site
    rows.Scan(&st.Name, &st.Login, &st.Pass, &st.Link)
    sites = append(sites, st)
  }
  return sites, nil
}

func loadDocs(dbpath string) ([]Doc, error) {
  docs := []Doc{}
  db, err := sql.Open("sqlite3", dbpath)
  if err != nil {
    return sites, fmt.Errorf("open %s db: %s", dbpath, err)
  }
  defer db.Close()
  query := "SELECT * FROM docs"
  rows, err := db.Query(query)
  if err != nil {
    return docs, fmt.Errorf("db query: %s", err)
  }
  defer rows.Close()
  for rows.Next() {
    var dc Doc
    rows.Scan(&dc.Name, &dc.Value)
    docs = append(docs, dc)
  }
  return docs, nil
}

func missingSites() ([]Site, int, int, error) {
  mst := []Site{}
  err := getRemoteDb()
  if err != nil {
    return mst, 0, 0, err
  }
  rems, err := loadSites("remote")
  if err != nil {
    return mst, 0, 0, err
  }
  fmt.Printf(" %d sites in remote DB\n", len(rems))
  locs, err := loadSites("wallt")
  if err!= nil {
    return mst, 0, 0, err
  }
  fmt.Printf(" %d sites in local DB\n", len(locs))
  Remote:
  for _, rst := range rems {
    for _, lst := range locs {
      if lst.Name == rst.Name {
        continue Remote
      }
    }
    mst = append(mst, rst)
  }
  return mst, len(rems), len(locs), nil
}

func missingDocs() ([]Site, int, int, error) {
  mdc := []Doc{}
  err := getRemoteDb()
  if err != nil {
    return mdc, 0, 0, err
  }
  rems, err := loadDocs("remote")
  if err != nil {
    return mdc, 0, 0, err
  }
  fmt.Printf(" %d documents in remote DB\n", len(rems))
  locs, err := loadDocs("wallt")
  if err!= nil {
    return mdc, 0, 0, err
  }
  fmt.Printf(" %d documents in local DB\n", len(locs))
  Remote:
  for _, rdc := range rems {
    for _, ldc := range locs {
      if ldc.Name == rdc.Name {
        continue Remote
      }
    }
    mdc = append(mdc, rdc)
  }
  return mdc, len(rems), len(locs), nil
}

func joinSites(sts []Site) error {
  db, err := sql.Open("sqlite3", locdb)
  if err != nil {
    return fmt.Errorf("open wallt.db: %s", err)
  }
  defer db.Close()
  cmd := `INSERT INTO sites VALUES (?, ?, ?, ?)`
  n := 0
  for i, s := range sts {
    _, er := db.Exec(cmd, s.Name, s.Login, s.Pass, s.Link)
    if er != nil {
      return fmt.Errorf("when insert into db: %s", er)
    }
    n = i
  }
  wd := "s were"
  if n == 1 {
    wd = " was"
  }
  fmt.Printf(" %d missing site%s added \n", n, wd)
  return nil
}

func joinDocs(ds []Doc) error {
  db, err := sql.Open("sqlite3", locdb)
  if err != nil {
    return fmt.Errorf("open wallt.db: %s", err)
  }
  defer db.Close()
  cmd := `INSERT INTO docs VALUES (?, ?)`
  n := 0
  for i, d := range ds {
    _, er := db.Exec(cmd, d.Name, d.Value)
    if er != nil {
      return fmt.Errorf("when insert into db: %s", er)
    }
    n = i
  }
  wd := "s were"
  if n == 1 {
    wd = " was"
  }
  fmt.Printf(" %d missing doc%s added \n", n, wd)
  return nil
}

func DoJoinDb() error {
  deal := false
  missts, nr, nl, err := missingSites()
  if err != nil {
    return fmt.Errorf("missingSites: %s", err)
  }
  if len(missts) != 0 {
    err = joinSites(missts)
    if err != nil {
      return fmt.Errorf("joinSites: %s", err)
    }
    if len(missts) + nl > nr {
      deal = true
    }
  }
  misdcs, nrs, nls, err := missingDocs()
  if err != nil {
    return fmt.Errorf("missingDocs: %s", err)
  }
  if len(misdcs) != 0 {
    err = joinDocs(misdcs)
    if err != nil {
      return fmt.Errorf("joinDocs: %s", err)
    }
    if len(misdcs) + nls > nrs {
      deal = true
    }
  }
  if deal {
    return UploadDb()
  }
  return nil
}

func UploadDb() error {
  auth, err := goph.Key(keypath, "")
  if err != nil {
    return fmt.Errorf("key: %s", err)
  }
  client, err := goph.New(user, raddr, auth)
  if err != nil {
    return fmt.Errorf("goph client: %s", err)
  }
  err = client.Upload(locdb, rdbpath)
  if err != nil {
    return fmt.Errorf("client upload: %s", err)
  }
  fmt.Println(" Database was uploaded")
  return os.Remove(tempath)
}
