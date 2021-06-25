package mysmsmasking

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

type client struct {
	url  string
	user string
	pass string
}

type Client interface {
	GetAccountInfo() (AccountInfo, error)
	GetReport(airwaybillId string) (Status, error)
	Send(msisdn, message string) (Airwaybill, error)
}

func (c client) Send(msisdn, message string) (Airwaybill, error) {
	rgxMsisdn := regexp.MustCompile(`^(0|62)8[1-9]\d+$`)
	if !rgxMsisdn.MatchString(msisdn) {
		return Airwaybill{}, errors.New("msisdn must begin with 628 or 08")
	}

	target := fmt.Sprintf(
		"%s/masking/send.php?username=%s&password=%s&hp=%s&message=%s",
		c.url,
		url.QueryEscape(c.user),
		url.QueryEscape(c.pass),
		msisdn,
		url.QueryEscape(message),
	)

	resp, err := http.Get(target)
	if err != nil {
		return Airwaybill{}, err
	}
	defer resp.Body.Close()
	respTime := time.Now()

	bytesResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return Airwaybill{}, err
	}

	return Airwaybill{
		Id:        string(bytesResp),
		Timestamp: respTime,
	}, nil
}

func (c client) GetReport(airwaybillId string) (Status, error) {
	target := fmt.Sprintf("%s/masking/report.php?rpt=%s", c.url, url.QueryEscape(airwaybillId))
	resp, err := http.Get(target)
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

func (c client) GetAccountInfo() (AccountInfo, error) {
	escUser := url.QueryEscape(c.user)
	escPass := url.QueryEscape(c.pass)
	target := fmt.Sprintf("%s/masking/balance.php?username=%s&password=%s", c.url, escUser, escPass)

	resp, err := http.Get(target)
	if err != nil {
		return AccountInfo{}, err
	}
	defer resp.Body.Close()

	bytesResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return AccountInfo{}, err
	}

	strResponse := rgxCSVSeparator.Split(string(bytesResp), -1)

	floatBal, err := strconv.ParseFloat(strResponse[0], 64)
	if err != nil {
		return AccountInfo{}, err
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return AccountInfo{}, err
	}

	expired, err := time.ParseInLocation("2006/01/02-15:04:05", strResponse[1], loc)
	if err != nil {
		return AccountInfo{}, err
	}

	return AccountInfo{
		Balance: uint32(floatBal),
		Expiry:  expired,
	}, nil
}

func New(serverurl, username, password string) Client {
	return &client{
		url:  serverurl,
		user: username,
		pass: password,
	}
}
