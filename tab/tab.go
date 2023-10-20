package main

import (
	"strings"
	"os"
	"bufio"
	"regexp"
	"fmt"
)

func main() {
	ph := "/data/data/com.termux/files/home/golangs/wallt/Makefile"
	f, err := os.Open(ph)
	if err != nil {
		fmt.Println(err)
	}
	txt := ""
	re := regexp.MustCompile(`^\s{2}\w.`)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
	  ln := sc.Text()
	  if re.MatchString(ln) {
	  	txt += strings.Replace(ln, "  ", "\t", 1) + "\n"
	  } else {
	  	txt += ln + "\n"
	  }
	}
	f.Close()
  err = os.WriteFile(ph, []byte(txt), 0640)
  if err != nil {
  	fmt.Println(err)	 
  }
}