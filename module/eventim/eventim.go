package module_eventim

import (
	"encoding/json"
	"errors"
	module "eventim-acc-gen/module"
	"eventim-acc-gen/requests"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/goombaio/namegenerator"
	uuid "github.com/satori/go.uuid"
)

type TestStructure struct {
	*module.TestStructure
}

var (
	proxy          = "http://PROXY"
	userAgent      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36"
	sensorGenLoops = 5
)

func (t *TestStructure) GenerateEventimAcc() error {

	t.Client = new(http.Client)
	t.Client = requests.CreateRequestClient(proxy)
	t.Host = "www.eventim.de"

	t.Headers.UserAgent = userAgent
	t.Headers.SecUa = `"Chromium";v="116", "Not)A;Brand";v="24", "Google Chrome";v="116"`
	t.Headers.Platform = `"macOS"`

	if err := t.visitHome(); err != nil {
		return err
	}

	if err := t.visitAkamaiUrl(); err != nil {
		return err
	}

	if t.Akamai.PixelActive {
		if err := t.visitPixelPage(); err != nil {
			return err
		}
	}

	if err := t.genInHouseSensorData(); err != nil {
		return err
	}

	if err := t.postSensorData(); err != nil {
		return err
	}

	if t.Akamai.PixelActive {
		if err := t.genInHousePixelData(); err != nil {
			return err
		}

		if err := t.postPixelData(); err != nil {
			return err
		}
	}

	for n := 0; n < sensorGenLoops; n++ {
		abck, _ := t.FindCookie("_abck")
		if strings.Contains(abck, "~0~") {
			break
		}
		if err := t.genInHouseSensorData(); err != nil {
			return err
		}

		if err := t.postSensorData(); err != nil {
			return err
		}
	}

	if err := t.visitCustomerData(); err != nil {
		return err
	}

	for n := 0; n < sensorGenLoops; n++ {
		abck, _ := t.FindCookie("_abck")
		if strings.Contains(abck, "~0~") {
			break
		}
		if err := t.genInHouseSensorData(); err != nil {
			return err
		}

		if err := t.postSensorData(); err != nil {
			return err
		}
	}

	if err := t.putRegistrationDetails(); err != nil {
		return err
	}

	return nil
}

func (t *TestStructure) visitHome() error {
	log.Println("visiting home")
	Request := requests.DoRequest{
		Client:         t.Client,
		CTX:            t.Context,
		AcceptedStatus: []int{200},
		Req: map[string]string{
			"Method": "GET",
			"URL":    fmt.Sprintf("https://%s/", t.Host),
			"Data":   "nil",
		},
		Headers: map[string][]string{
			"Sec-Ch-Ua":          {t.Headers.SecUa},
			"Sec-Ch-Ua-Platform": {t.Headers.Platform},
			"Sec-Ch-Ua-Mobile":   {"?0"},
			"User-Agent":         {t.Headers.UserAgent},
			"accept":             {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
			"sec-fetch-site":     {"none"},
			"sec-fetch-mode":     {"navigate"},
			"sec-fetch-user":     {"?1"},
			"sec-fetch-dest":     {"document"},
			"accept-encoding":    {"gzip, deflate, br"},
			"accept-language":    {"en-US,en;q=0.9"},
		},
	}
	response := Request.MakeRequest()
	if response.Error != nil {
		log.Println("error visiting home =>", response.Error.Error())
		return response.Error
	}
	log.Println("visiting home =>", response.RespStatus)
	t.Headers.Referrer = Request.Req["URL"]
	if strings.Contains(response.ResponseBody, "bazadebe") {
		t.Akamai.PixelActive = true
		bv, pt, pv, err := ParseHomePage(response.ResponseBody)
		if err != nil {
			return errors.New(err.Err)
		}
		t.Akamai.T = between(pt, "=", "&")
		t.Akamai.Baza = bv
		t.Akamai.PixelV = pv
	}

	if t.Akamai.AkamWebURL == "" {
		AkamWebURL := between(response.ResponseBody, "<noscript><img src=\"https://"+t.Host, "</html>")
		t.Akamai.AkamWebURL = fmt.Sprintf("https://%s%s", t.Host, between(AkamWebURL, "src=\"", "\"></script></body>"))
	}

	// Scrape for static web if empty
	if t.Akamai.AkamWebURL == "" {
		t.Akamai.AkamWebURL = between(response.ResponseBody, "_cf.push(['_setAu', '", "']); ")
	}

	t.Akamai.AkamWebURL = "https://www.eventim.de/W7oU/XGtI/E-/Jb-d/Mjug/EpOJtmak5f/HUBaJAE/cgtZG/XxlSzsB"

	log.Println("sensor url =>", t.Akamai.AkamWebURL)
	log.Println("pixel version =>", t.Akamai.PixelV)
	return nil
}

func (t *TestStructure) visitAkamaiUrl() error {
	log.Println("getting akamai sensor url")
	Request := requests.DoRequest{
		Client:         t.Client,
		CTX:            t.Context,
		AcceptedStatus: []int{200, 201},
		Req: map[string]string{
			"Method": "GET",
			"URL":    t.Akamai.AkamWebURL,
			"Data":   "nil",
		},
		Headers: map[string][]string{
			"accept":             {"*/*"},
			"sec-ch-ua":          {t.Headers.SecUa},
			"sec-ch-ua-platform": {t.Headers.Platform},
			"user-agent":         {t.Headers.UserAgent},
			"Sec-Fetch-Site":     {"same-origin"},
			"Sec-Fetch-Mode":     {"no-cors"},
			"Sec-Fetch-Dest":     {"script"},
			"referrer":           {t.Headers.Referrer},
			"accept-encoding":    {"gzip, deflate, br"},
			"accept-language":    {"en-US,en;q=0.9"},
		},
	}
	response := Request.MakeRequest()
	if response.Error != nil {
		log.Println("error getting akamai sensor url =>", response.Error.Error())
		return response.Error
	}
	log.Println("getting akamai sensor url =>", response.RespStatus)
	switch response.RespStatus {
	case 200, 201:
		return nil
	default:
		return errors.New("error getting akamai sensor url")
	}
}

func (t *TestStructure) visitPixelPage() error {
	log.Println("getting pixel url")
	Request := requests.DoRequest{
		Client:         t.Client,
		CTX:            t.Context,
		AcceptedStatus: []int{200, 201},
		Req:            map[string]string{"Method": "GET", "URL": fmt.Sprintf("https://%s/akam/13/%s", t.Host, t.Akamai.PixelV), "Data": "nil"},
		Headers:        map[string][]string{"accept": {"*/*"}, "sec-ch-ua": {t.Headers.SecUa}, "sec-ch-ua-platform": {t.Headers.Platform}, "user-agent": {t.Headers.UserAgent}, "Sec-Fetch-Site": {"same-origin"}, "Sec-Fetch-Mode": {"no-cors"}, "Sec-Fetch-Dest": {"script"}, "referrer": {t.Headers.Referrer}, "accept-encoding": {"gzip, deflate, br"}, "accept-language": {"en-US,en;q=0.9"}},
	}
	response := Request.MakeRequest()
	if response.Error != nil {
		log.Println("error getting pixel url =>", response.Error.Error())
		return response.Error
	}
	log.Println("getting pixel url =>", response.RespStatus)
	switch response.RespStatus {
	case 200, 201:
		t.Akamai.ScriptVal, _ = ParsePixelScript(response.ResponseBody)
		return nil
	default:
		return errors.New("error getting the pixel page")
	}
}

func (t *TestStructure) visitCustomerData() error {
	log.Println("visiting customer data")
	Request := requests.DoRequest{
		Client:         t.Client,
		CTX:            t.Context,
		AcceptedStatus: []int{200},
		Req: map[string]string{
			"Method": "GET",
			"URL":    fmt.Sprintf("https://%s/mycustomerdata/", t.Host),
			"Data":   "nil",
		},
		Headers: map[string][]string{
			"Sec-Ch-Ua":          {t.Headers.SecUa},
			"Sec-Ch-Ua-Platform": {t.Headers.Platform},
			"Sec-Ch-Ua-Mobile":   {"?0"},
			"User-Agent":         {t.Headers.UserAgent},
			"accept":             {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
			"sec-fetch-site":     {"none"},
			"sec-fetch-mode":     {"navigate"},
			"sec-fetch-user":     {"?1"},
			"sec-fetch-dest":     {"document"},
			"referrer":           {t.Headers.Referrer},
			"accept-encoding":    {"gzip, deflate, br"},
			"accept-language":    {"en-US,en;q=0.9"},
		},
	}
	response := Request.MakeRequest()
	if response.Error != nil {
		log.Println("error visiting customer data =>", response.Error.Error())
		return response.Error
	}
	log.Println("visiting customer data =>", response.RespStatus)
	t.Headers.Referrer = Request.Req["URL"]
	return nil
}

func (t *TestStructure) putRegistrationDetails() error {
	log.Println("registering new account")
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	postData := map[string]interface{}{
		"contactData": []map[string]interface{}{
			{
				"fieldName": "customerSalutation",
				"value":     "Frau",
			},
			{
				"fieldName": "customerFirstName",
				"value":     strings.Split(nameGenerator.Generate(), " ")[0],
			},
			{
				"fieldName": "customerLastName",
				"value":     strings.Split(nameGenerator.Generate(), " ")[0],
			},
			{
				"fieldName": "customerCompany",
				"value":     "",
			},
			{
				"fieldName": "customerStreetAndNo",
				"value":     fmt.Sprintf("%.f Main St", math.Floor(999)),
			},
			{
				"fieldName": "customerPostalCode",
				"value":     "08902",
			},
			{
				"fieldName": "customerCity",
				"value":     "North Brunswick",
			},
			{
				"fieldName": "customerCountry",
				"value":     "US",
			},
			{
				"fieldName": "customerEmail",
				"value":     fmt.Sprintf("%s@gmail.com", uuid.NewV1().String()),
			},
			{
				"fieldName": "customerPhone",
				"value":     fmt.Sprintf("732%.f", math.Floor(9999999)),
			},
			{
				"fieldName": "customerMobile",
				"value":     "",
			},
			{
				"fieldName": "customerDateOfBirth",
				"value":     "",
			},
		},
		"password": "H46VeP7nPPx6Tru!",
	}
	postJson, _ := json.Marshal(postData)
	Request := requests.DoRequest{
		Client:         t.Client,
		CTX:            t.Context,
		AcceptedStatus: []int{200, 201},
		Req:            map[string]string{"Method": "PUT", "URL": "https://www.eventim.de/api/customers/register/?force_session=true&affiliate=EVE", "Data": string(postJson)},
		Headers: map[string][]string{
			"accept":             {"application/json, text/plain, */*"},
			"content-type":       {"application/json"},
			"x-csrf-token":       {"C382FAD13C364FDB84855D468E04F4C4"},
			"sec-ch-ua":          {t.Headers.SecUa},
			"sec-ch-ua-platform": {t.Headers.Platform},
			"user-agent":         {t.Headers.UserAgent},
			"Sec-Fetch-Site":     {"same-origin"},
			"Sec-Fetch-Mode":     {"cors"},
			"Sec-Fetch-Dest":     {"empty"},
			"referrer":           {t.Headers.Referrer},
			"origin":             {fmt.Sprintf("https://%s", t.Host)},
			"accept-encoding":    {"gzip, deflate, br"},
			"accept-language":    {"en-US,en;q=0.9"},
		},
	}
	response := Request.MakeRequest()
	if response.Error != nil {
		log.Println("error registering new account =>", response.Error.Error())
		return response.Error
	}
	log.Println("posting registering new account =>", response.RespStatus)
	return nil
}
