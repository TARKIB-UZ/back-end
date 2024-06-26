package webapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

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

func (a *AuthWebAPI) SendSMSWithAndroid(ctx context.Context, phoneNumber string, code string) error {
	secret := os.Getenv("SECRET_SMS_GATEWAY")
	device := os.Getenv("SMS_ANDROID_DEVICE_ID")
	mode := "devices"

	url := "https://sms.uncgateway.com/api/send/sms"

	message := "Your OTP code is: " + code

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("secret", secret)
	_ = writer.WriteField("mode", mode)
	_ = writer.WriteField("phone", phoneNumber)
	_ = writer.WriteField("message", message)
	_ = writer.WriteField("device", device)
	_ = writer.WriteField("sim", "2")

	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("SMS sent successfully!")
		return nil
	} else {
		fmt.Println("Failed to send SMS. Status code:", resp.StatusCode)
		return errors.New(resp.Status)
	}
}
