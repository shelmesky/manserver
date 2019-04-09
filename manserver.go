package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

var (
	manPath    = flag.String("man", "/usr/bin/man", "man path of system")
	zcatPath   = flag.String("zcat", "/usr/bin/zcat", "zcat path of system")
	groffPath  = flag.String("groff", "/usr/bin/groff", "groff path of system")
	bashPath   = flag.String("bash", "/bin/bash", "bash path of system")
	listenAddr = flag.String("listen", "0.0.0.0:9876", "tcp listen address and port")
)

func SearchMan(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "form.html")
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		keyword := r.FormValue("keyword")
		keyword = strings.Trim(keyword, " ")

		// 根据关键字搜索man文档的路径(只显示一个结果， -f选项可以搜索书多个结果， 但不显示路径)
		manSearchCommand := fmt.Sprintf("%s -w %s", *manPath, keyword)

		cmd := exec.Command(*bashPath, "-c", manSearchCommand)
		var stdout, stderr bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			log.Println("cmd.Run() man -w failed with %s\n", err)
			fmt.Fprintf(w, string(stderr.Bytes()))
		}

		errStr := string(stderr.Bytes())
		if stderr.Len() > 0 {
			fmt.Println("exec failed:", errStr)
		}

		// 使用zcat打开man文档路径并使用groff转化为html格式
		docPath := strings.Trim(string(stdout.Bytes()), "\n")
		openMandocCommand := fmt.Sprintf("%s %s | %s -Thtml -mandoc", *zcatPath, docPath, *groffPath)
		openMandocCmd := exec.Command(*bashPath, "-c", openMandocCommand)

		var openMandocStdout, openMandocStderr bytes.Buffer

		openMandocCmd.Stdout = &openMandocStdout
		openMandocCmd.Stderr = &openMandocStderr

		err = openMandocCmd.Run()
		if err != nil {
			log.Println("cmd.Run() zcat failed with %s\n", err)
			fmt.Fprintf(w, string(stderr.Bytes()))
		}

		w.Write(openMandocStdout.Bytes())

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func main() {
	flag.Parse()

	http.HandleFunc("/", SearchMan)

	fmt.Printf("Starting linux man document server at %s \n", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatal(err)
	}
}
