package main

import (
    "fmt"
    "database/sql"
    "time"
    "errors"
    "os"
    "io"
    "os/exec"
    "github.com/yurajp/wallt/conf"
    _ "github.com/mattn/go-sqlite3"
    "github.com/yurajp/wallt/purecrypt"
    "github.com/melbahja/goph"
)

var (
  tempath = "data/remote.db"
  locdb = "data/wallt.db"
  user string
  raddr string
  rdbpath string
  keypath string
)

/// TODO Add docs & Passport
func (app *App) RecodeDb(newWord string) error {
    nDb, err := sql.Open("sqlite3", "data/temp.db")
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
    err = os.Rename("data/temp.db", "data/wallt.db")
    if err != nil {
      return err
    }
    db, err := sql.Open("sqlite3", "data/wallt.db")
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
  fdb, err := os.Open("data/wallt.db")
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

//// DELETE FUNC
func (app *App) ShareDb() error {
  wd, _ := os.Getwd()
  upld := "/storage/emulated/0/Uploads/"
  cmd := exec.Command("cp", "data/wallt.db", upld)
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
    return docs, fmt.Errorf("open %s db: %s", dbpath, err)
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

func missingDocs() ([]Doc, int, int, error) {
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
  err := conf.GetRemoteCfg()
  if err != nil {
    return fmt.Errorf("remote config: %s", err)
  }
  user = conf.RemoteCfg.User
  raddr = conf.RemoteCfg.Addr
  rdbpath = conf.RemoteCfg.RDbPath
  keypath = conf.RemoteCfg.KeyPath

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
  } else {
    fmt.Println(" No sence to share database")
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

