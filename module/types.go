package module

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

type TestStructure struct {
	Context  context.Context
	Client   *http.Client
	Host     string
	Headers  Headers
	Response map[string]interface{}
	Akamai   Akamai
}

type Headers struct {
	UserAgent string
	SecUa     string
	Platform  string
	Referrer  string
}

type Akamai struct {
	Sensor *InhouseSensorAkamai
	Pixel  *InhousePixelAkamai
	Sec
}

type InhouseSensorAkamai struct {
	ApiKey     string `json:"apiKey"`
	Ua         string `json:"ua"`
	PageURL    string `json:"pageUrl"`
	ApiVersion string `json:"apiVersion"`
	Abck       string `json:"_abck"`
	BmSz       string `json:"bm_sz"`
	Hash       string `json:"hash"`
}

type InhousePixelAkamai struct {
	ApiKey    string `json:"apiKey"`
	Ua        string `json:"ua"`
	ScriptVal string `json:"scriptVal"`
	PixelID   string `json:"pixelId"`
}

type Sec struct {
	PixelActive bool
	GotSession  bool
	AkamWebURL  string
	Baza        string
	PixelV      string
	T           string
	ScriptVal   string
	PixelData   string
	SensorData  string
	ScriptMD5   string
}

// Add a cookie to the current cookie jar
func (t *TestStructure) AddCookie(name, value, domain, host string) {
	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:   name,
		Value:  value,
		Path:   "/",
		Domain: domain,
	}
	cookies = append(cookies, cookie)
	u, _ := url.Parse(host)
	t.Client.Jar.SetCookies(u, cookies)
}

func (t *TestStructure) FindCookie(name string) (string, error) {
	u, _ := url.Parse("https://" + t.Host)
	for _, v := range t.Client.Jar.Cookies(u) {
		if v.Name == name {
			return v.Value, nil
		}
	}
	return "", errors.New("cookie does not exist: " + name)
}
