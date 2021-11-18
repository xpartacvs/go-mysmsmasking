package client

import (
	"errors"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/xpartacvs/go-mysmsmasking/core"
)

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
	core core.Core
}

const (
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

	ErrInvalidMSISDN error = errors.New("msisdn must begin with 628 or 08")
)

func (c Client) GetAccountInfo() (*AccountInfo, error) {
	resp, err := c.core.CallApi(core.MethodGet, "masking", "balance", url.Values{})
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

	resp, err := c.core.CallApi(core.MethodPost, "masking", "send", data)
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

	resp, err := c.core.CallApi(core.MethodGet, "masking", "report", data)
	if err != nil {
		return FAILED, err
	}
	defer resp.Body.Close()

	bytesResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return FAILED, err
	}

	strBody := string(bytesResp)

	rgxCVS := regexp.MustCompile(`^\d+.*`)
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

func New(username, password string) Client {
	return Client{
		core: core.New(username, password),
	}
}
