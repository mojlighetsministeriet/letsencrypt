package main // import "github.com/mojlighetsministeriet/letsencrypt"

import (
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/mojlighetsministeriet/utils"
)

// This service should be given the external URL of https://yourdomainname.com/.well-known

func registerCertificateForIfNew(domainName string) (err error) {
	if _, statError := os.Stat("/etc/letsencrypt/renewal/" + domainName + ".conf"); !os.IsNotExist(statError) {
		return
	}

	cmd := exec.Command("certbot", "certonly", "--webroot", "-w /var/www", "-d "+domainName)
	err = cmd.Run()

	return
}

func replaceSecret(name string, value string) (err error) {
	cmd := exec.Command("docker", "secret", "rm", name)
	cmd.Run()

	cmd = exec.Command("docker", "secret", "create", name, value)
	err = cmd.Run()

	return
}

func updateTLSSecrets(domainNames []string, logger echo.Logger) {
	for {
		for _, domainName := range domainNames {
			content, err := ioutil.ReadFile("/etc/live/" + domainName + "/fullchain.pem")
			if err != nil {
				logger.Error(err)
			}

			err = replaceSecret("tls-cert-"+domainName, string(content))
			if err != nil {
				logger.Error(err)
			}

			content, err = ioutil.ReadFile("/etc/live/" + domainName + "/privkey.pem")
			if err != nil {
				logger.Error(err)
			}

			err = replaceSecret("tls-key-"+domainName, string(content))
			if err != nil {
				logger.Error(err)
			}
		}

		time.Sleep(time.Minute)
	}
}

func main() {
	domainNameInput := os.Getenv("DOMAINS")
	if domainNameInput == "" {
		panic("The environment variable DOMAINS must contain a comma separated string of at least one domain name.")
	}

	domainNames := regexp.MustCompile("\\s*,\\s*").Split(domainNameInput, -1)

	service := echo.New()
	service.Logger.SetLevel(log.INFO)

	service.Use(middleware.Gzip())
	service.Use(middleware.Static("/var/www/.well-known"))

	for _, domainName := range domainNames {
		err := registerCertificateForIfNew(domainName)
		if err != nil {
			panic(err)
		}
	}

	go updateTLSSecrets(domainNames, service.Logger)

	panic(service.Start(":" + utils.GetEnv("PORT", "80")))
}
