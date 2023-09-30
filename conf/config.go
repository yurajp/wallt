package conf

import (
  "fmt"
  "errors"
  "encoding/json"
  "os"
  "time"
  "regexp"
  "strconv"
)

type Config struct {
  Port string `json: "port"`
  Livetime time.Duration `json: "livetime"`
  Appdir string `json:"appdir"`
}

type RemoteConfig struct {
  User string `json:"user"`
  Addr string `json:"addr"`
  RDbPath string `json:"rdbpath"`
  KeyPath string `json:"keypath"`
}

var Cfg Config
var RemoteCfg RemoteConfig

func ConfigExists() bool {
  _, err := os.Stat("conf/Config.json")
  if os.IsNotExist(err) {
    return false
  }
  return true
}

func SetConfigTerm() *Config {
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
    fmt.Println("\n  Print time in minutes\n  when password is valid\n  or enter for default (3 min)")
    fmt.Scanf("%s", &t)
    if t == "" {
      valtime = "3"
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
  appdir, err := os.Getwd()
  if err != nil {
    fmt.Println(err)
    return &Config{}
  }
  return &Config{port, livetime, appdir}
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

func WriteConfig(cfg *Config) error {
  js, err := json.Marshal(cfg)
  if err != nil {
    return err
  }
  return os.WriteFile("conf/Config.json", js, 0640)
}

func GetConfig() error {
  if !ConfigExists() {
    return errors.New("Config does not exist")
  }
  js, err := os.ReadFile("conf/Config.json")
  if err != nil {
    return err
  }
  return json.Unmarshal(js, &Cfg)
}

func Prepare() error {
  if ConfigExists() {
    return nil
  }
  cfg := SetConfigTerm()
  return WriteConfig(cfg)
}

func SetRemoteConf() {
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
  rcf := RemoteConfig{u, a, p, k}
  jrc, _ := json.Marshal(rcf)
  os.WriteFile("conf/RemoteConf.json", jrc, 0640)
}

func GetRemoteCfg() error {
  _, err := os.Stat("conf/RemoteConf.json")
  if os.IsNotExist(err) {
    SetRemoteConf()
  }
  js, err := os.ReadFile("conf/RemoteConf.json")
  if err != nil {
    return err
  }
  return json.Unmarshal(js, &RemoteCfg)
}