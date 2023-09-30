package models

import (
  "regexp"
  "strconv"
  
)


func checkNum(n string) bool {
  re := regexp.MustCompile(`\d{4}\s?\d{4}\s?\d{4}\s?\d{4}`)
  return re.MatchString(n)
}  

func checkCvc(n string) bool {
  re := regexp.MustCompile(`\d\d\d`)
  return re.MatchString(n)
}

func checkDate(my []string) bool {
  if len(my) != 2 {
    return false
  }
  re := regexp.MustCompile(`\d\d`)
  if !re.MatchString(my[0]) || !re.MatchString(my[1]) {
    return false
  }
  dm, _ := strconv.Atoi(my[0])
  dy, _ := strconv.Atoi(my[1])
  if dm < 1 || dm > 12 || dy < 23 || dy > 35 {
    return false
  }
  return true
}
