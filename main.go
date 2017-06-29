package main

import (
	"bufio"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/docopt/docopt-go"
	cfg "gopkg.in/gcfg.v1"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
)

type Document struct {
	Issuer     string
	IssuerName string
	Recipient  string
	Value      string
	ValueWords string
	Date       string
}

type Profile struct {
}

type Config struct {
	General struct {
		TemplatePath string
		LogLevel     log.Level
	}
	Profile map[string]*Profile
}

const (
	//TODO: Set template path
	templatePath = "dist/usr/share/quitty/zuwendungsbestatigung_geldzuwendung.tex"
	configPath   = "dist/etc/quitty.conf"
)

func main() {
	usage := `quitty

Usage:
  quitty    new <recipient> <value> <date>
  quitty    -h | --help

Options:
  -h --help    Show this screen.
  --version    Print version information.
`
	arguments, err := docopt.Parse(usage, nil, true, "0.1", false)
	if err != nil {
		log.WithField("Error", err).Fatalln("Error parsing arguments")
	}

	config := Config{}
	err = cfg.ReadFileInto(&config, configPath)
	if err != nil {
		log.WithField("Error", err).Fatalln("Error reading config")
	}
	log.Info("Config read successfull")

	log.WithField("arguments", arguments).Debug("read cli arguments")
	log.WithField("config", config).Debug("read config values")

	pdflatex_path, err := exec.LookPath("pdflatex")
	if err != nil {
		log.WithField("Error", err).
			Fatalln("pdflatex seems not to be in $PATH\nPlease check your LaTeX installation")
	}

	pdflatex_version_cmd := exec.Command(pdflatex_path, "--version")
	pdflatex_version_stdout, err := pdflatex_version_cmd.StdoutPipe()
	if err != nil {
		log.WithField("Error", err).
			Fatalln("could not open stdout from inferior process")
	}

	go func() {
		scanner := bufio.NewScanner(pdflatex_version_stdout)
		for scanner.Scan() {
			log.WithField("pdflatex version", scanner.Text()).Debug()
		}
	}()

	err = pdflatex_version_cmd.Run()
	if err != nil {
		log.WithField("Error", err).
			Fatalln("could not retrieve version information from pdflatex")
	}

	tmpdir, err := ioutil.TempDir("", "quitty")
	if err != nil {
		log.WithField("Error", err).
			Fatalln("could not create temporary directory")
	}
	log.WithField("path", tmpdir).Debug("created temporary directory")

	log.RegisterExitHandler(func() {
		err := os.RemoveAll(tmpdir)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
				"Path":  tmpdir,
			}).
				Fatalln("could not clean temporary directory")
		}
		log.WithField("path", tmpdir).Debug("deleted temporary directory")
	})

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.WithField("Error", err).
			Fatalln("could not parse template")
	}
	tmpl = tmpl.Delims("<<", ">>")

	pdflatex_executer_cmd := exec.Command(pdflatex_path, fmt.
		Sprintf("--output-directory=%v --jobname=quitty", tmpdir))

	stdout_pipe, err := pdflatex_executer_cmd.StdoutPipe()
	if err != nil {
		log.WithField("Error", err).
			Fatalln("could not open stdout from inferior process")
	}

	stderr_pipe, err := pdflatex_executer_cmd.StderrPipe()
	if err != nil {
		log.WithField("Error", err).
			Fatalln("could not open stderr from inferior process")
	}

	go func() {
		scanner := bufio.NewScanner(stdout_pipe)
		for scanner.Scan() {
			log.WithField("pdflatex", scanner.Text()).Info()
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr_pipe)
		for scanner.Scan() {
			log.WithField("pdflatex", scanner.Text()).Error()
		}
	}()

	err = pdflatex_executer_cmd.Run()
	if err != nil {
		log.WithField("Error", err).
			Fatalln("pdflatex returned an error")
	}
}
