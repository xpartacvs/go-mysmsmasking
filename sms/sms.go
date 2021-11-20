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
	"sync"
	"time"
)

type allowedMethod string

// Status menunjukan status dari suatu kiriman SMS.
type Status int8

// AccountInfo berisi info terkair akun mysmsmasking
// yang terdiri dari info saldo dan tanggal kedaluarsa.
type AccountInfo struct {
	Balance int64
	Expiry  time.Time
}

// Airwaybill beisi info terkait kiriman SMS yang
// terdisi dari Id kiriman dan sempel waktu kapan SMS
// disubmit ke gateway MySMSMasking.
type Airwaybill struct {
	Id        string
	Timestamp time.Time
}

// Client adalah struct utama yang akan berinteraksi
// dengan API MySMSMasking.
type Client struct {
	user string
	pass string
}

const (
	defaultBaseUrl string = "http://sms.mysmsmasking.com"

	methodGet  allowedMethod = http.MethodGet
	methodPost allowedMethod = http.MethodPost

	FAILED               Status = iota // Pengiriman SMS gagal
	SENT                               // SMS sudah dikirimkan dari server MySMSMasking ke MSISDN tujuan
	DELIVERED                          // SMS sudah diterima olah MSISDN tujuan
	INVALID_ID                         // Airwaybill tidak valid
	INVALID_MSISDN                     // Nomor MSISDN tidak valid
	BALANCE_INSUFFICIENT               // Saldo habis/kurang
	BALANCE_EXPIRED                    // Akun MySMSMasking kedaluarsa
)

var (
	// Error karena nomor MSISDN tidak valid
	ErrInvalidMSISDN error = errors.New("msisdn must begin with 628 or 08")

	rgxCSVSeparator *regexp.Regexp = regexp.MustCompile(`,\s*`)
	rgxMsisdn       *regexp.Regexp = regexp.MustCompile(`^(0|62)8[1-9]\d+$`)
	rgxCVS          *regexp.Regexp = regexp.MustCompile(`^\d+.*`)

	errMethodNotAllowed error = errors.New("method not allowed")

	runtimeBaseUrl string
	once           sync.Once
)

func runtimeBaseURL() string {
	once.Do(func() {
		r, err := url.ParseRequestURI(os.Getenv("MYSMSMASKING_BASEURL"))
		if err != nil {
			runtimeBaseUrl = defaultBaseUrl
			return
		}

		r.RawQuery = ""
		r.RawFragment = ""
		r.Fragment = ""

		if len(r.String()) <= 0 {
			runtimeBaseUrl = defaultBaseUrl
			return
		}

		runtimeBaseUrl = strings.TrimRight(r.String(), "/ ")
	})

	return runtimeBaseUrl
}

func (c Client) callApi(method allowedMethod, namespace, apiName string, data url.Values) (*http.Response, error) {
	urlPath := runtimeBaseURL() + "/" + namespace + "/" + apiName + ".php"

	data.Add("username", c.user)
	data.Add("password", c.pass)

	switch method {
	case methodGet:
		return http.Get(urlPath + "?" + data.Encode())
	case methodPost:
		return http.PostForm(urlPath, data)
	}

	return nil, errMethodNotAllowed
}

// Jika tidak terjadi error, method GetAccountInfo akan mengembalikan *AccountInfo.
// Sebaliknya, error akan berisi bukan nil.
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

// Jika tidak terjadi error, method Send akan mengembalikan *Airwaybill.
// Sebaliknya, error akan berisi bukan nil.
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

// Jika tidak terjadi error, method GetStatus akan mengembalikan
// salah satu konstanta dengan type Status. Sebaliknya, error
// akan berisi bukan nil.
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

// NewClient adalah func untuk menginstansiasi type Client
// dengan memanfaatkan arguman username dan password.
func NewClient(username, password string) Client {
	return Client{
		user: username,
		pass: password,
	}
}
