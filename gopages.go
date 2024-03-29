// Copyright 2010 Abiola Ibrahim <abiola89@gmail.com>. All rights reserved.
// Use of this source code is governed by New BSD License
// http://www.opensource.org/licenses/bsd-license.php
// The content and logo is governed by Creative Commons Attribution 3.0
// The mascott is a property of Go governed by Creative Commons Attribution 3.0
// http://creativecommons.org/licenses/by/3.0/

package main

import (
	"bufio"
	"code.google.com/p/gopages/util"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

const (
	MAKE    = iota
	GOBUILD = iota
)

//where the build execution starts
func main() {
	cl := flag.Bool("clean", false, "don't build, just clean the generated pages")
	//run := flag.Bool("run", false, "run the generated executable after build")
	flag.Parse()
	if *cl {
		err := clean()
		if err != nil {
			println(err.Error())
		}
		return
	}
	settings, err := util.LoadSettings() //inits the settings and generates the .go source files
	if err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}
	util.Config = settings.Data //stores settings to accessible variable
	println("generated", len(settings.Data["pages"]), "gopages")
	println()
	err = util.AddHandlers(settings.Data["pages"]) //add all handlers
	if err != nil {
		println(err.Error())
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "get" {
		build("")
	} else if len(os.Args) > 1 {
		println("unrecognized ", "'"+os.Args[1]+"'")
		println("  gopages get - build project with go get after generating pages")
	} else {
		build("pages")
		println("run \"gopages get\" to build project with go get after generating pages")
	}
}

//create the pages directory to store generated source codes
func init() {
	err := os.MkdirAll(util.DIR, 0755)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

//to build the project with gobuild or make after generating .go source files
func build(folder string) (err error) {
	tmp := path.Join(os.TempDir(), "gopagestmp")
	tmpFile, err := os.OpenFile(tmp, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	fd := []*os.File{os.Stdin, tmpFile, os.Stderr}
	goexec, _ := exec.LookPath("go")
	if len(goexec) == 0 {
		return errors.New("go not found in PATH")
	}
	dir := os.Getenv("PWD")
	if folder != "" {
		dir = path.Join(dir, folder)
	}
	//	id, err := os.ForkExec(gofmt, []string{"", "-w", file}, os.Environ(), dir, fd)
	process, err := os.StartProcess(goexec, []string{"", "get"}, &os.ProcAttr{Env: os.Environ(), Files: fd, Dir: dir})
	if err != nil {
		return
	} else {
		process.Wait()
		checkerr(tmp, folder)
		//os.Wait(id, 0)
	}
	return
}

//check compile errors
func checkerr(tmp string, folder string) {
	t, _ := ioutil.ReadFile(util.PATHS_FILE)
	ls := strings.Fields(string(t))
	paths := make(map[string]string)
	for i := 0; i < len(ls); i += 2 {
		if folder == "" {
			paths[ls[i]] = ls[i+1]
		} else if folder == "pages" {
			paths[strings.Replace(ls[i], "pages", ".", 1)] = ls[i+1]
		}
	}
	file, err := os.Open(tmp)
	if err != nil {
		return
	}
	out := ""
	buf := bufio.NewReader(file)
outer:
	for {
		l, _, err := buf.ReadLine()
		if err != nil {
			break
		}
		line := string(l)
		out += fmt.Sprintln(line)
		if strings.HasSuffix(line, "pages") {
			for {
				l, _, err := buf.ReadLine()
				if err != nil {
					break outer
				}
				line = string(l)
				s := strings.SplitN(line, ":", 3)
				if len(s) < 3 {
					out += fmt.Sprintln(line)
					continue
				}
				if str, ok := paths[s[0]]; ok {
					out += fmt.Sprintln(strings.Replace(line, s[0]+":"+s[1], str, 1))
				} else {
					out += fmt.Sprintln(line)
				}
			}
		}
	}
	println(out)
}

//deletes the generated source codes
func clean() (err error) {
	err = os.RemoveAll(util.DIR)
	if err != nil {
		println(err.Error())
	}
	return
}
