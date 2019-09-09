package main

import (
	"bytes"
	"disney-bookinger/config"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
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
エラーレポートは DM or LINE へ
`

func newSESClient(conf config.Config) (*ses.SES, error) {
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

func createQuery(reserve config.Reserve) url.Values {
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
	return queryAPI
}

func doCheckReservation(reserve config.Reserve, conf config.Config) error {
	q := createQuery(reserve)
	fmt.Println(fmt.Sprintf("Checking: %s %s", reserve.Endpoint, q.Encode()))
	resp, err := http.PostForm(reserve.Endpoint, q)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(string(b))
	}
	body := string(b)
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
	q := createQuery(reserve)
	fmt.Println(fmt.Sprintf("Checking: %s %s", reserve.WeeklyEndpoint, q.Encode()))
	resp, err := http.Post(reserve.WeeklyEndpoint,
		"application/x-www-form-urlencoded; charset=UTF-8",
		bytes.NewReader([]byte(q.Encode())),
	)
	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	body := string(b)
	if !strings.Contains(body, "空席あり") {
		return nil
	}
	dom, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
	if err != nil {
		return err
	}

	// per 1day (<- 1week)
	dom.Find("table > tbody > tr").Each(func(_ int, s1 *goquery.Selection) {
		if !strings.Contains(s1.Text(), "空席あり") {
			return
		}
		s1.Find("tr > td > a > ul").Each(func(_ int, s2 *goquery.Selection) {
			if !strings.Contains(s2.Find("li > p.sheet").Text(), "S席") {
				return
			}
			if strings.Contains(s2.Find("li > p.state").Text(), "満席") {
				return
			}
			if strings.Contains(s2.Find("li > p.state").Text(), "予約不可") {
				return
			}

			rawDate, _ := s1.Find("th > span.display").Html()
			queryLink := url.Values{}
			flLink := reflect.Indirect(reflect.ValueOf(reserve.Link))
			ftLink := flLink.Type()

			for i := 0; i < ftLink.NumField(); i++ {
				queryLink.Add(ftLink.Field(i).Tag.Get("json"), flLink.Field(i).String())
			}

			d, _ := time.Parse("20060102", rawDate)
			queryLink.Set("useDate", rawDate)
			date := d.Format("2006/01/02")

			link := reserve.Pagelink + queryLink.Encode()
			msg := fmt.Sprintf(template, reserve.Title, date, link)
			err = sendMail(conf, msg, reserve.Title, date)
			if err != nil {
				reportError(conf, err)
			}
		})
	})
	return nil
}

func checkWeeklyReservation(reserve config.Reserve, conf config.Config) {
	wg := sync.WaitGroup{}
	for i := 0; i <= 4 ; i++ {
		wg.Add(1)
		go func(res config.Reserve, idx int) {
			res.API.UseDate = time.Now().Add(time.Hour * 24 * time.Duration(idx) * 7).Format("20060102")
			res.Link.UseDate = res.API.UseDate
			defer wg.Done()
			err := doCheckWeeklyReservation(res, conf)
			if err != nil {
				reportError(conf, err)
			}
		}(reserve, i)
	}
	wg.Wait()
}

func sendMail(conf config.Config, message, title, date string) error {
	sesClient, err := newSESClient(conf)
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
	sesClient, _ := newSESClient(conf)
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
		if reserve.IsWeeklyCheck {
			checkWeeklyReservation(reserve, *conf)
			if err != nil {
				reportError(*conf, err)
			}
			continue
		}
		err := doCheckReservation(reserve, *conf)
		if err != nil {
			reportError(*conf, err)
		}
	}
}