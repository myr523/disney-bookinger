package main

import (
	"disney-bookinger/config"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

const (
	CONFPATH = "./config/config.json"
)

var template = `
ディズニー空席確認Botからのお知らせだよ:
%s %s 空席あり
予約ページのリンク -> %s

from disney-bookinger @ no__2_
エラーレポートは DMへ
`

func NewSESClient(conf config.Config) (*ses.SES, error) {
	cfgs := &aws.Config{
		Region:   aws.String("us-east-1"),
		Endpoint: aws.String("email.us-east-1.amazonaws.com"),
		Credentials: credentials.NewStaticCredentials(
			conf.AWSSESConfig.AWSSESKey,
			conf.AWSSESConfig.AWSSESSecret,
			"",
		),
	}
	sess, err := session.NewSession(cfgs)
	if err != nil {
		return nil, err
	}
	sesClient := ses.New(sess)
	return sesClient, nil
}

func doCheckReservation(reserve config.Reserve, conf config.Config) error {
	queryAPI := url.Values{}
	flAPI := reflect.Indirect(reflect.ValueOf(reserve.API))
	ftAPI := flAPI.Type()

	for i := 0; i < ftAPI.NumField(); i++ {
		tag := ftAPI.Field(i).Tag
		if (string(tag) == "showId" || string(tag) == "seatRank") && !reserve.IsShowRestaurant {
			continue
		}
		queryAPI.Add(tag.Get("json"), flAPI.Field(i).String())
	}
	fmt.Println(fmt.Sprintf("Checking: %s %s", reserve.Endpoint, queryAPI.Encode()))
	resp, err := http.PostForm(reserve.Endpoint, queryAPI)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	body := string(b)
	fmt.Println(body)
	if strings.Contains(body, "空席あり") {
		queryLink := url.Values{}
		flLink := reflect.Indirect(reflect.ValueOf(reserve.Link))
		ftLink := flLink.Type()

		for i := 0; i < ftLink.NumField(); i++ {
			queryLink.Add(ftLink.Field(i).Tag.Get("json"), flLink.Field(i).String())
		}

		d, _ := time.Parse("20060102", reserve.Link.UseDate)
		date := d.Format("2006/01/02")
		link := reserve.Pagelink + queryLink.Encode()
		msg := fmt.Sprintf(template, reserve.Title, date, link)
		err := sendMail(conf, msg, reserve.Title, date)
		if err != nil {
			return err
		}
	}
	return nil
}

func doCheckWeeklyReservation(reserve config.Reserve, conf config.Config) error {
	return nil
}

func sendMail(conf config.Config, message, title, date string) error {
	sesClient, err := NewSESClient(conf)
	if err != nil {
		return err
	}
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: conf.SendTo,
			BccAddresses: aws.StringSlice([]string{conf.MailDebug}),
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text:&ses.Content{
					Charset: aws.String("UTF-8"),
					Data: aws.String(message),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data: aws.String(fmt.Sprintf("【%s】空席情報: %s", date, title)),
			},
		},
		Source: aws.String(conf.MailFrom),
	}
	_, err = sesClient.SendEmail(input)
	if err != nil {
		return err
	}
	return nil
}

func reportError(conf config.Config, er error) {
	sesClient, _ := NewSESClient(conf)
	fmt.Println(er)
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(conf.MailDebug),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(er.Error()),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String("error info"),
			},
		},
		Source: aws.String(conf.MailDebug),
	}
	_, _ = sesClient.SendEmail(input)
}

func main() {
	conf, err := config.Parse(CONFPATH)
	if err != nil {
		panic(err)
	}
	for _, reserve := range conf.Reserves {
		err := doCheckReservation(reserve, *conf)
		if err != nil {
			reportError(*conf, err)
		}
	}
}