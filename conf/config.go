package conf

import (
  "fmt"
//  "path/filepath"
//  "errors"
//  "encoding/json"
  "os"
  "time"
  "regexp"
  "strconv"
  "github.com/yurajp/confy"
)

type Config struct {
  Port string
  Livetime time.Duration
  Appdir string
  Remote RemoteConfig
}

type RemoteConfig struct {
  User string
  Addr string
  RDbPath string
  KeyPath string
}

var (
	Cfg Config
	Path = "/data/data/com.termux/files/home/golangs/wallt/conf/Config.ini"
)


func ConfigExists() bool {
  _, err := os.Stat(Path)
  if os.IsNotExist(err) {
    return false
  }
  return true
}

func SetConfigTerm() Config {
  var port, valtime string
  done := false
  for !done {
    fmt.Print("\n  Print port number\n  or enter for default (8686)\n  ")
    var p string
    fmt.Scanf("%s", &p)
    if p == "" {
      port = ":8686"
      done = true
    } else {
      if PortIsCorrect(p) {
        port = ":" + p
        done = true
      } else {
        fmt.Println(" Port should be number between \n  1024 and 49151")
      }
    }
  }
  done = false
  var t string
  for !done {
    fmt.Println("\n  Print time in minutes\n  when password is valid\n  or enter for default (5 min)")
    fmt.Scanf("%s", &t)
    if t == "" {
      valtime = "5"
      done = true
    } else {
      if TimeIsCorrect(t) {
        valtime = t
        done = true
      } else {
        fmt.Println("  Time must be number")
      }
    }
  }
  livetime, _ := time.ParseDuration(valtime + "m")
  here, err := os.Getwd()
  if err != nil {
    fmt.Println(err)
    return Config{}
  }
//  appdir := filepath.Dir(here)
  appdir := here
  fmt.Printf(" APPDIR is %s\n enter to continue\n", appdir)
  var cte string
  fmt.Scanf("%s", &cte)

  rcf := SetRemoteConf()
  return Config{port, livetime, appdir, rcf}
}

func PortIsCorrect(p string) bool {
  re := regexp.MustCompile(`^\d\d\d\d(\d)?$`)
  if !re.MatchString(p) {
    return false
  }
  d, _ := strconv.Atoi(p)
  if d < 1024 || d > 49151 {
    return false
  }
  return true
}

func TimeIsCorrect(t string) bool {
  re := regexp.MustCompile(`^\d{1,}$`)
  if !re.MatchString(t) {
    return false
  }
  return true
}


func GetConfig() error {
  c := Config{}
  confy.Path = Path
  intf, err := confy.LoadConfy(c)
  if err != nil {
    return err
  }
  Cfg = intf.(Config)
  return nil
}

func Prepare() error {
  if ConfigExists() {
    return nil
  }
  cfg := SetConfigTerm()
  confy.Path = Path
  return confy.WriteConfy(cfg)
}

func SetRemoteConf() RemoteConfig {
  fmt.Println(" Remote user name")
  var u string
  fmt.Scanf("%s", &u)
  fmt.Println(" Remote server addres")
  var a string
  fmt.Scanf("%s", &a)
  fmt.Println(" Path to wallt db on remote machine")
  var p string
  fmt.Scanf("%s", &p)
  fmt.Println(" Path to local public SSH key")
  var k string
  fmt.Scanf("%s", &k)
  return RemoteConfig{u, a, p, k}
  
}

