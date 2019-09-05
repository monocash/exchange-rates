package exchanger

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// @link : https://themoneyconverter.com/AED/EUR.aspx?amount=1

type theMoneyConverterAPI struct {
	attributes
}

var (
	theMoneyConverterAPIURL     = `https://themoneyconverter.com/%s/%s.aspx?amount=1`
	theMoneyConverterAPIHeaders = map[string][]string{
		`Accept`: {`text/html`},
	}
)

func (c *theMoneyConverterAPI) requestRate(from string, to string, opt ...interface{}) (*theMoneyConverterAPI, error) {

	// todo add option opt to add more headers or client configurations
	// free mem-leak
	// optimize for memory leak
	// todo optimize curl connection

	// format the url and replace currency
	url := fmt.Sprintf(theMoneyConverterAPIURL, from, to)
	// prepare the request
	req, _ := http.NewRequest("GET", url, nil)
	// assign the request headers
	theMoneyConverterAPIHeaders[`User-Agent`] = []string{c.userAgent}
	req.Header = theMoneyConverterAPIHeaders

	// execute the request
	res, err := c.Client.Do(req)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// free mem-leak
	// todo discard data
	c.responseBody = string(body)
	return c, nil
}

// GetRateValue ... get exchange rate value
func (c *theMoneyConverterAPI) GetRateValue() float64 {
	return c.rateValue
}

// GetRateDateTime ... return rate datetime
func (c *theMoneyConverterAPI) GetRateDateTime() string {
	return c.rateDate.Format(time.RFC3339)
}

// GetExchangerName ... return exchanger name
func (c *theMoneyConverterAPI) GetExchangerName() string {
	return c.name
}

// Latest ... populate latest exchange rate
func (c *theMoneyConverterAPI) Latest(from string, to string, opt ...interface{}) error {

	// todo cache layer
	_, err := c.requestRate(from, to, opt)
	if err != nil {
		log.Print(err)
		return err
	}

	// if from currency is same as converted currency return value of 1
	if from == to {
		c.rateValue = 1
		return nil
	}

	validID := regexp.MustCompile(`(?s)output(.*)>(.*)</output>`)
	stringMatches := validID.FindStringSubmatch(c.responseBody)

	stringMatch := strings.TrimSpace(strings.Replace(stringMatches[2], "\n", "", -1))
	stringMatch = strings.Replace(stringMatch, fmt.Sprintf("%d %s = ", 1, from), "", -1)
	stringMatch = strings.Replace(stringMatch, fmt.Sprintf(" %s", to), "", -1)

	c.rateValue, err = strconv.ParseFloat(stringMatch, 64)
	c.rateDate = time.Now()

	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

// NewtheMoneyConverterAPI ... return new instance of theMoneyConverterAPI
func NewtheMoneyConverterAPI(opt map[string]string) *theMoneyConverterAPI {

	keepAliveTimeout := 600 * time.Second
	timeout := 5 * time.Second
	defaultTransport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: keepAliveTimeout,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	}

	client := &http.Client{
		Transport: defaultTransport,
		Timeout:   timeout,
	}

	attr := attributes{
		name:      `themoneyconverter`,
		Client:    client,
		userAgent: `Mozilla/5.0 (Macintosh; Intel Mac OS X 10.8; rv:21.0) Gecko/20100101 Firefox/21.0`,
	}
	if opt[`userAgent`] != "" {
		attr.userAgent = opt[`userAgent`]
	}

	r := &theMoneyConverterAPI{attr}

	return r
}
