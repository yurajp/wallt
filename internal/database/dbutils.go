package database

import (
    "fmt"
    "database/sql"
    "time"
    "errors"
    "os"
    "io"
    
    "github.com/yurajp/wallt/conf"
    "github.com/yurajp/wallt/internal/models"
    _ "github.com/mattn/go-sqlite3"
    "github.com/yurajp/wallt/internal/purecrypt"
    "github.com/melbahja/goph"
)

var (
  tempath = "../data/remote.db"
  locdb = "../data/wallt.db"
  user = conf.RemoteCfg.User
  raddr = conf.RemoteCfg.Addr
  rdbpath = conf.RemoteCfg.RDbPath
  keypath = conf.RemoteCfg.KeyPath
)


func (wdb *Wdb) RecodeDb(word, newWord string) error {
    nDb, err := sql.Open("sqlite3", "../data/temp.db")
    if err != nil {
       return err
    }
    defer nDb.Close()
    nWdb:= &Wdb{nDb}
    nWdb.CreateTables()
    
    querySs := `select * from sites`
    queryA := `insert in sites(name, login, pass, link) values(?, ?, ?, ?)`
    rowsSs, err := wdb.Db.Query(querySs)
    if err != nil {
      return err
    }
    for rowsSs.Next() {
      var nm, lg, ps, lk string
      rowsSs.Scan(&nm, &lg, &ps, &lk)
      nlg := purecrypt.Symcode(purecrypt.Desymcode(lg, word), newWord)
      nps := purecrypt.Symcode(purecrypt.Desymcode(ps, word), newWord)
      _, err = nDb.Exec(queryA, nm, nlg, nps, lk)
      if err != nil {
        return err
      }
    }
    queryCs := `select * from cards`
    queryB := `insert in cards(name, number, expire, cvc) values(?, ?, ?, ?)`
    rowsCs, err := wdb.Db.Query(queryCs)
    if err != nil {
      return err
    }
    for rowsCs.Next() {
      var nc, nb, ex, cv string
      rowsCs.Scan(&nc, &nb, &ex, &cv)
      nnb := purecrypt.Symcode(purecrypt.Desymcode(nb, word), newWord)
      nex := purecrypt.Symcode(purecrypt.Desymcode(ex, word), newWord)
      ncv := purecrypt.Symcode(purecrypt.Desymcode(cv, word), newWord)
      _, err = nDb.Exec(queryB, nc, nnb, nex, ncv)
      if err != nil {
        return err
      }
    }
    
    queryPp := `select * from passrf`
    queryC := `insert into passrf(serialnum, date, whom, code) values(?, ?, ?, ?)`
    rowsPp, err := wdb.Db.Query(queryPp)
    if err != nil {
      return err
    }
    var nsn, ndt, nwh, ncd string
    var sn, dt, wh, cd string
    for rowsPp.Next() {
      rowsPp.Scan(&sn, &dt, &wh, &cd)
      nsn = purecrypt.Symcode(purecrypt.Desymcode(sn, word), newWord)
      ndt = purecrypt.Symcode(purecrypt.Desymcode(dt, word), newWord)
      nwh = purecrypt.Symcode(purecrypt.Desymcode(wh, word), newWord)
      ncd = purecrypt.Symcode(purecrypt.Desymcode(cd, word), newWord)
    }
    _, err = nDb.Exec(queryC, nsn, ndt, nwh, ncd)
    if err != nil {
      return err
    }
    
    queryDs := `select * from docs`
    queryD := `insert in docs(name, value) values(?, ?)`
    rowsDs, err := wdb.Db.Query(queryDs)
    if err != nil {
      return err
    }
    for rowsDs.Next() {
      var dn, dv string
      rowsCs.Scan(&dn, &dv)
      ndn := purecrypt.Symcode(purecrypt.Desymcode(dn, word), newWord)
      ndv := purecrypt.Symcode(purecrypt.Desymcode(dv, word), newWord)
      _, err = nDb.Exec(queryD, ndn, ndv)
      if err != nil {
        return err
      }
    }
    nDb.Close()
    wdb.Db.Close()
    err = os.Rename("../data/temp.db", "../data/wallt.db")
    if err != nil {
      return err
    }
    err = purecrypt.WriteCheckword(newWord)
    if err != nil {
      return err
    }
    return nil
}


func BackupDb() error {
  ty, tm, td := time.Now().Date()
  date := fmt.Sprintf("%v%v%v", ty - 2000, tm, td)
  i, err := os.Stat("../archive")
  if errors.Is(err, os.ErrNotExist) || !i.IsDir() {
    err = os.Mkdir("../archive", 0775)
    if err != nil {
      return fmt.Errorf(" Cannot make dir:\n%s", err)
    }
  }
  fpath := fmt.Sprintf("../archive/%s.wallt.db", date)
  f, err := os.Create(fpath)
  if err != nil {
    return err
  }
  defer f.Close()
  fdb, err := os.Open(locdb)
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

func loadSites(dbpath string) ([]models.Site, error) {
  sites := []models.Site{}
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
    var st models.Site
    rows.Scan(&st.Name, &st.Login, &st.Pass, &st.Link)
    sites = append(sites, st)
  }
  return sites, nil
}

func loadDocs(dbpath string) ([]models.Doc, error) {
  docs := []models.Doc{}
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
    var dc models.Doc
    rows.Scan(&dc.Name, &dc.Value)
    docs = append(docs, dc)
  }
  return docs, nil
}

func missingSites() ([]models.Site, int, int, error) {
  mst := []models.Site{}
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

func missingDocs() ([]models.Doc, int, int, error) {
  mdc := []models.Doc{}
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

func joinSites(sts []models.Site) error {
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

func joinDocs(ds []models.Doc) error {
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

