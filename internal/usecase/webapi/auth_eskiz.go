package webapi

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/k0kubun/pp"
	"tarkib.uz/config"
)

type AuthWebAPI struct {
	cfg *config.Config
}

func NewAuthWebAPI(cfg *config.Config) *AuthWebAPI {
	return &AuthWebAPI{
		cfg: cfg,
	}
}

func (a *AuthWebAPI) SendSMS(ctx context.Context, phoneNumber string, code string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("mobile_phone", phoneNumber)
	_ = writer.WriteField("message", "This is test from Eskiz")
	_ = writer.WriteField("from", "tarkib.uz")

	err := writer.Close()
	if err != nil {
		pp.Println(err)
		return err
	}

	// Ensure the URL does not have surrounding quotes
	apiEndpoint := a.cfg.SMS.APIEndpoint
	apiEndpoint = apiEndpoint[1 : len(apiEndpoint)-1] // Remove leading and trailing quotes

	req, err := http.NewRequest("POST", apiEndpoint, body)
	if err != nil {
		pp.Println(err)
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", a.cfg.SMS.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		pp.Println(err)
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		pp.Println(err)
		return err
	}

	pp.Println(string(respBody))

	return nil
}
