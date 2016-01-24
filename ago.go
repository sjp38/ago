// Copyright (c) 2016 SeongJae Park.
//
// This program is free software; you can redistribute it and/or modify it
// under the terms of the GNU General Public License version 3 as published by
// the Free Software Foundation.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
)

const (
	CMD          = "ago"
	USAGE        = "USAGE: " + CMD + " <commands> [argument ...]\n"
	NOARG_ERRMSG = USAGE + "\nFor detail, try " + CMD + " help\n"
	HELP_MSG     = "Use the source ;)\n"
	ANDRD        = "android"
	ANDRD_TMPDIR = "/data/local/tmp"
	DOCSDIR      = "docs"
	DOCINFO      = "info"
	WORDINFO     = "words"
	DOCDIR_PREF  = "doc" // prefix of document directory
)

var (
	errl        = log.New(os.Stderr, "[err] ", 0)
	dbgl        = log.New(os.Stderr, "[dbg] ", 0)
	metadat_dir = "/tmp/.ago"
	docs_dir    string
	doci_path   string // documents information file path
	wordi_path  string // words information file path
	docs_info   documents_info
)

// document contains informations for a document.
type document struct {
	Name string
	Id   int // id of the document.
}

type documents_info struct {
	Docs    []document
	Next_id int
}

func read_docs_info() {
	c, err := ioutil.ReadFile(doci_path)
	if err != nil {
		errl.Printf("failed to read doc info file: %s\n", err)
		os.Exit(1)
	}

	if err := json.Unmarshal(c, &docs_info); err != nil {
		errl.Printf("error while unmarshal doc info: %s\n", err)
		dbgl.Printf("the json: %s\n", c)
		os.Exit(1)
	}
}

func write_docs_info() {
	bytes, err := json.Marshal(docs_info)
	if err != nil {
		errl.Printf("failed to marshal docs_info: %s\n", err)
		os.Exit(1)
	}

	if err := ioutil.WriteFile(doci_path, bytes, 0600); err != nil {
		errl.Printf("failed to write marshaled docs_info: %s\n", err)
	}
}

func lsdocs(args []string) {
	for _, doc := range docs_info.Docs {
		fmt.Printf("%d: %s\n", doc.Id, doc.Name)
	}
}

// file_exists checks whether a file of specific path exists.
//
// Returns true if exists, false if not.
func file_exists(path string) bool {
	if _, err := os.Stat(docs_dir); err == nil {
		return true
	}
	return false
}

func analyze_words(bytes []byte) {
	// TODO: analyze words in the document and save the information
	fmt.Printf("analyze...\n%s\n", bytes)
}

func adddoc(file_path string) error {
	if !file_exists(file_path) {
		err := errors.New("file not exists")
		return err
	}

	// read the file
	bytes, err := ioutil.ReadFile(file_path)
	if err != nil {
		msg := fmt.Sprintf("failed to read file: %s", err)
		err := errors.New(msg)
		return err
	}

	// analyze words in the file content
	analyze_words(bytes)

	// create dir under docs/
	docid := docs_info.Next_id
	docdir := fmt.Sprintf("%s%d", DOCDIR_PREF, docid)
	docdirpath := path.Join(docs_dir, docdir)
	if err = os.MkdirAll(docdirpath, 0700); err != nil {
		msg := fmt.Sprintf("failed to create dir: %s", err)
		err = errors.New(msg)
		return err
	}

	// write copy in the doc<doc id>/
	_, docname := filepath.Split(file_path)
	in_file_path := path.Join(docdirpath, docname)
	if err = ioutil.WriteFile(in_file_path, bytes, 0600); err != nil {
		msg := fmt.Sprintf("failed to write file: %s", err)
		err = errors.New(msg)
		return err
	}

	// add to docs_info global object
	doc := document{Name: docname, Id: docid}
	docs_info.Next_id += 1
	docs_info.Docs = append(docs_info.Docs, doc)
	return nil
}

func adddocs(args []string) {
	for _, path := range args {
		if err := adddoc(path); err != nil {
			errl.Printf("failed to add doc %s: %s\n",
				path, err)
			os.Exit(1)
		}
	}
	write_docs_info()
}

// rmdoc remove a document with specific id.
//
// Returns true if success to remove the document, false if not
func rmdoc(target int) error {
	for idx, doc := range docs_info.Docs {
		if doc.Id != target {
			continue
		}
		docdir := fmt.Sprintf("%s%d", DOCDIR_PREF, doc.Id)
		docdirpath := path.Join(docs_dir, docdir)
		if err := os.RemoveAll(docdirpath); err != nil {
			msg := fmt.Sprintf("failed to remove dir: %s", err)
			err = errors.New(msg)
			return err
		}

		docs_info.Docs = append(docs_info.Docs[:idx],
			docs_info.Docs[idx+1:]...)
		return nil
	}
	return errors.New("no such doc")
}

func rmdocs(args []string) {
	for _, arg := range args {
		target, err := strconv.Atoi(arg)
		if err != nil {
			errl.Printf("argument must be doc id. err: \n", err)
		}
		if err = rmdoc(target); err != nil {
			fmt.Printf("failed to remove doc id %d: %s\n",
				target, err)
		}
	}

	write_docs_info()
}

func do_test(args []string) {
	fmt.Printf("do test %s\n", args)
}

// main is the entry point of `ago`.
// ago usage is similar to familiar tools:
// 	ago <command> [argument ...]
//
// commands are:
// ls-docs, add-docs, rm-docs: list, add, remove documentation[s].
// doc, mod-doc: Commands for future. Not be implemented yet. Display and
// modify information of the doc.
// test: start a test. Number of questions can be specified as option.
//
// The description above is lie because this program is nothing for now. It is
// just a plan.
func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("No argument.\n")
		fmt.Printf(NOARG_ERRMSG)
		os.Exit(1)
	}
	cmd := args[1]
	args = args[2:]
	switch cmd {
	case "ls-docs":
		lsdocs(args)
	case "add-docs":
		adddocs(args)
	case "rm-docs":
		rmdocs(args)
	case "test":
		do_test(args)
	case "help":
		fmt.Printf(HELP_MSG)
	default:
		errl.Printf("wrong commanad")
		os.Exit(1)
	}
}

// init initializes few things for `ago`.
// Internally, ago uses a metadata directory for state saving. Path of the
// directory is `$HOME/.ago`. If `$HOME` is not exists, `/tmp` is used as
// default. For Android support in future, it should be `/data/local/tmp` at
// future.
//
// Hierarchy of the directory is as:
// .ago/docs/doc1
//          /info
//     /words
//
// Documents added by user resides under .ago/docs/ with its own directory. The
// document own directories are named as doc[id] which id is an integer.
// Metadata about those documents are recorded under .ago/docs/info file. The
// metadata contains original document name and current location under .ago
// directory. Because current ago support only text file, this structure is
// unnecessary. Actually, the struct is for future scaling. In future, ago will
// be an document organizer like Mendeley[1] and will support not only text
// file, but also pdf, odt, url, etc.
//
// File `words` under the .ago/ directory contains all data for words in the
// documents. It contains each word and its frequency in the docs(in total and
// per each doc), score of user, and meaning of the word. Calculated importance
// can be in there maybe but not yet decided to add it.
//
// [1] https://www.mendeley.com/
func init() {
	if runtime.GOOS == ANDRD {
		metadat_dir = ANDRD_TMPDIR
	}

	if os.Getenv("HOME") != "" {
		metadat_dir = os.Getenv("HOME")
	}
	metadat_dir = path.Join(metadat_dir, ".ago")

	docs_dir = path.Join(metadat_dir, DOCSDIR)
	doci_path = path.Join(docs_dir, DOCINFO)
	wordi_path = path.Join(metadat_dir, WORDINFO)

	if file_exists(docs_dir) {
		read_docs_info()
		return
	}

	dbgl.Printf("docs dir is not exists. Create it.\n")
	err := os.MkdirAll(docs_dir, 0700)
	if err != nil {
		errl.Printf("docs dir %s creation failed: %s\n", docs_dir, err)
		os.Exit(1)
	}

	for _, file := range []string{doci_path, wordi_path} {
		f, err := os.Create(file)
		if err != nil {
			errl.Printf("docs info file creation failed: %s\n", err)
			os.Exit(1)
		}
		f.Close()

		if err = os.Chmod(file, 0600); err != nil {
			errl.Printf("chmod file %s failed: %s\n", err)
			os.Exit(1)
		}
	}

	docs_info.Next_id = 0
	write_docs_info()
	read_docs_info()
}
