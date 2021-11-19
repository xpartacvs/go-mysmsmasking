package sms

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Method string
type Status int8

type AccountInfo struct {
	Balance int64
	Expiry  time.Time
}

type Airwaybill struct {
	Id        string
	Timestamp time.Time
}

type Client struct {
	user string
	pass string
}

const (
	methodGet  Method = http.MethodGet
	methodPost Method = http.MethodPost

	FAILED Status = iota
	SENT
	DELIVERED
	INVALID_ID
	INVALID_MSISDN
	BALANCE_INSUFFICIENT
	BALANCE_EXPIRED
)

var (
	rgxCSVSeparator *regexp.Regexp = regexp.MustCompile(`,\s*`)
	rgxMsisdn       *regexp.Regexp = regexp.MustCompile(`^(0|62)8[1-9]\d+$`)
	rgxCVS          *regexp.Regexp = regexp.MustCompile(`^\d+.*`)

	ErrInvalidMSISDN    error = errors.New("msisdn must begin with 628 or 08")
	ErrMethodNotAllowed error = errors.New("method not allowed")
)

func (c Client) callApi(method Method, namespace, apiName string, data url.Values) (*http.Response, error) {
	baseUrl := "http://sms.mysmsmasking.com"

	if envBaseUrl := strings.Trim(os.Getenv("MYSMSMASKING_BASEURL"), "/ "); len(envBaseUrl) > 0 {
		baseUrl = envBaseUrl
	}

	urlPath := baseUrl + "/" + namespace + "/" + apiName + ".php"
	data.Add("username", c.user)
	data.Add("password", c.pass)

	switch method {
	case methodGet:
		return http.Get(urlPath + "?" + data.Encode())
	case methodPost:
		return http.PostForm(urlPath, data)
	}

	return nil, ErrMethodNotAllowed
}

func (c Client) GetAccountInfo() (*AccountInfo, error) {
	resp, err := c.callApi(methodGet, "masking", "balance", url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytesResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	strResponse := rgxCSVSeparator.Split(string(bytesResp), -1)

	floatBal, err := strconv.ParseFloat(strResponse[0], 64)
	if err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}

	expired, err := time.ParseInLocation("2006/01/02-15:04:05", strResponse[1], loc)
	if err != nil {
		return nil, err
	}

	return &AccountInfo{
		Balance: int64(floatBal),
		Expiry:  expired,
	}, nil
}

func (c Client) Send(msisdn, message string) (*Airwaybill, error) {
	if !rgxMsisdn.MatchString(msisdn) {
		return nil, ErrInvalidMSISDN
	}

	data := url.Values{}
	data.Add("hp", msisdn)
	data.Add("message", message)

	resp, err := c.callApi(methodPost, "masking", "send", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respTime := time.Now()

	bytesResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Airwaybill{
		Id:        string(bytesResp),
		Timestamp: respTime,
	}, nil
}

func (c Client) GetStatus(airwaybillId string) (Status, error) {
	data := url.Values{}
	data.Add("rpt", airwaybillId)

	resp, err := c.callApi(methodGet, "masking", "report", data)
	if err != nil {
		return FAILED, err
	}
	defer resp.Body.Close()

	bytesResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return FAILED, err
	}

	strBody := string(bytesResp)

	if !rgxCVS.MatchString(strBody) {
		return INVALID_ID, nil
	}

	slcResp := rgxCSVSeparator.Split(strBody, -1)
	if len(slcResp) <= 0 {
		return INVALID_ID, nil
	}

	// switch slcResp[0] {
	// case "50":
	// 	return BALANCE_INSUFFICIENT, nil
	// case "51":
	// 	return BALANCE_EXPIRED, nil
	// case "52":
	// 	return INVALID_MSISDN, nil
	// case "20":
	// 	return SENT, nil
	// case "22":
	// 	return DELIVERED, nil
	// default:
	// 	return INVALID_ID, nil
	// }

	switch slcResp[0] {
	case "52":
		fallthrough
	case "51":
		fallthrough
	case "50":
		return FAILED, nil
	case "20":
		return SENT, nil
	case "22":
		return DELIVERED, nil
	default:
		return INVALID_ID, nil
	}
}

func NewClient(username, password string) Client {
	return Client{
		user: username,
		pass: password,
	}
}
