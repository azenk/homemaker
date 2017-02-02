/*
 * Copyright (c) 2015 Alex Yatskov <alex@foosoft.net>
 * Author: Alex Yatskov <alex@foosoft.net>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"io/ioutil"
)

const (
	flagClobber = 1 << iota
	flagForce
	flagVerbose
	flagNoCmds
	flagNoLinks
	flagNoMacro
	flagUnlink = flagNoCmds | (1 << iota)
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] [conf] src\n", path.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "https://foosoft.net/projects/homemaker/\n\n")
	fmt.Fprintf(os.Stderr, "Parameters:\n")
	flag.PrintDefaults()
}

func main() {
	taskName := flag.String("task", "default", "name of task to execute")
	dstDir := flag.String("dest", os.Getenv("HOME"), "target directory for tasks")
	force := flag.Bool("force", true, "create parent directories to target")
	clobber := flag.Bool("clobber", false, "delete files and directories at target")
	verbose := flag.Bool("verbose", false, "verbose output")
	nocmds := flag.Bool("nocmds", false, "don't execute commands")
	nolinks := flag.Bool("nolinks", false, "don't create links")
	variant := flag.String("variant", "", "execution variant for tasks and macros")
	unlink := flag.Bool("unlink", false, "remove existing links instead of creating them")

	flag.Usage = usage
	flag.Parse()

	flags := 0
	if *clobber {
		flags |= flagClobber
	}
	if *force {
		flags |= flagForce
	}
	if *verbose {
		flags |= flagVerbose
	}
	if *nocmds {
		flags |= flagNoCmds
	}
	if *nolinks {
		flags |= flagNoLinks
	}
	if *unlink {
		flags |= flagUnlink
	}

	var confFile string
	var srcDir string
	if flag.NArg() == 2 {
		confFile = makeAbsPath(flag.Arg(0))
		srcDir = makeAbsPath(flag.Arg(1))
	} else if flag.NArg() == 1 {
		// default to using a tasks.yml file if found in the specified directory, otherwise no config
		srcDir = makeAbsPath(flag.Arg(0))
		defaultConfFilePath := path.Join(srcDir,"tasks.yml")
		if _, err := os.Stat(defaultConfFilePath); err == nil {
			confFile = defaultConfFilePath
		}
	} else {
		usage()
		os.Exit(2)
	}

	var conf *config
	if confFile != "" {
		var err error
		conf, err = newConfig(confFile)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// if no config is found,
		// assume all files/folders found should be links
		conf = &config{handled: make(map[string]bool), Tasks: make(map[string]task)}
		default_task := task{}

		entries, err := ioutil.ReadDir(srcDir)
		if err != nil { log.Fatalf("Unable to build list of links: %s", err) }
		for _, entry := range entries {
			default_task.Links = append(default_task.Links, []string{entry.Name(), entry.Name()})
		}
		conf.Tasks["default"] = default_task
	}

	conf.srcDir = makeAbsPath(flag.Arg(1))
	conf.dstDir = makeAbsPath(*dstDir)
	conf.variant = *variant
	conf.flags = flags

	os.Setenv("HM_CONFIG", confFile)
	os.Setenv("HM_TASK", *taskName)
	os.Setenv("HM_SRC", conf.srcDir)
	os.Setenv("HM_DEST", conf.dstDir)
	os.Setenv("HM_VARIANT", conf.variant)

	if err := processTask(*taskName, conf); err != nil {
		log.Fatal(err)
	}
}
