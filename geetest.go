package geetest

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/parnurzeal/gorequest"
)

const apiHost = "api.geetest.com"

// Geetest is used for captcha registration and validation
type Geetest struct {
	captchaID  string
	privateKey string

	registerURL, validateURL string
}

// New constructs and returns a Geetest
func New(captchaID, privateKey string, enableHTTPS bool) Geetest {
	scheme := "http"
	if enableHTTPS {
		scheme = "https"
	}
	apiServer := fmt.Sprintf("%s://%s", scheme, apiHost)

	return Geetest{
		captchaID:  captchaID,
		privateKey: privateKey,

		registerURL: fmt.Sprintf("%s/register.php", apiServer),
		validateURL: fmt.Sprintf("%s/validate.php", apiServer),
	}
}

// CaptchaID returns captchaID
func (g Geetest) CaptchaID() string { return g.captchaID }

// Register returns challenge get from api server
func (g Geetest) Register() (string, error) {
	query := struct {
		CaptchaID string `json:"gt"`
	}{g.captchaID}
	_, body, errs := gorequest.New().Get(g.registerURL).Query(query).Timeout(time.Second * 2).End()
	if errs != nil {
		return "", &multierror.Error{Errors: errs}
	}

	return hexmd5(body + g.privateKey), nil
}

// Validate validates challenge
func (g Geetest) Validate(challenge, validate, seccode string) (bool, error) {
	hash := g.privateKey + "geetest" + challenge
	if validate != hexmd5(hash) {
		return false, nil
	}

	query := struct {
		Seccode string `json:"seccode"`
	}{seccode}
	_, body, errs := gorequest.New().Post(g.validateURL).Query(query).End()
	if errs != nil {
		return false, &multierror.Error{Errors: errs}
	}

	return hexmd5(seccode) == body, nil
}

func hexmd5(data string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}
