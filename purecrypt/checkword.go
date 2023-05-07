package purecrypt

import (
  "fmt"
  "os"
  "encoding/json"
  "errors"
)

type Checkword struct {
  Enc string `json: "enc"`
}

func ChWordExists() bool {
  _, err := os.Stat("checkword.json")
  if os.IsNotExist(err) {
    return false
  }
  return true
}

func WriteCheckword(pw string) error {
  if pw == "" {
    return errors.New("Empty password")
  }
  phr := "Password is Correct"
  chw := Checkword{Symcode(phr, pw)}
  f, err := os.Create("checkword.json")
  if err != nil {
    return err
  }
  defer f.Close()
  jsw, err := json.Marshal(chw)
  if err != nil {
    return err
  }
  _, err = f.Write(jsw)
  if err != nil {
    return err
  }
  return nil
}

func IsCorrect(pw string) bool {
  f, err := os.Open("checkword.json")
  if err != nil {
    return false
  }
  defer f.Close()
  var chw Checkword
  err = json.NewDecoder(f).Decode(&chw)
  if err != nil {
    fmt.Println(err)
    return false
  }
  phr := "Password is Correct"
  if Desymcode(chw.Enc, pw) == phr {
    return true
  }
  return false
}
