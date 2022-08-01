package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

const (
	defaultPermisions = 0777
)

var (
	text           string
	timePeriod     string
	storageDir     string
	responseLogDir string
	wait           time.Duration
)

func init() {
	flag.StringVar(&text, "text", "", "That will be serched in yandex images")
	flag.StringVar(&timePeriod, "tbs", "", `
	Set the period for which the image was published
	If you want imagies in all period just don't enter this flag
	Params:
		d - in 24 hours period
		w - in week period
		m - in month period
		y - in year period
	`)
	flag.StringVar(&storageDir, "img-storage", "storage", "Path for saving scrapped images links")
	flag.StringVar(&responseLogDir, "resp-log", "log", "Path for saving responses in txt format")
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	if text == "" {
		log.Panic(`
			Please enter text, that will be serched in yandex images.
		For more information execute program with -? flag.
		`)
	}
}

func main() {
	flag.Parse()
	d := make(chan os.Signal, 1)
	signal.Notify(d, os.Interrupt)

	waiter := sync.WaitGroup{}

	err := os.MkdirAll(storageDir, defaultPermisions)
	if err != nil {
		log.Panicln(err)
	}
	err = os.MkdirAll(responseLogDir, defaultPermisions)
	if err != nil {
		log.Panicln(err)
	}

	scrpUrl := configureUrl()
	c := colly.NewCollector()

	c.OnHTML(`img.yWs4tf`, func(e *colly.HTMLElement) {
		fmt.Println(e.Attr("src"))
		waiter.Add(1)
		go func() {
			err := saveInfo(string(e.Response.Body), e.Attr("src"))
			if err != nil {
				log.Println(err)
			}
			waiter.Done()
		}()
	})

	// переход на следующую страницу для скраппинга
	c.OnHTML(`a.frGj1b`, func(e *colly.HTMLElement) {
		select {
		case <-d:
			waiter.Wait()
			os.Exit(0)
		default:
			e.Request.Visit(e.Attr("href"))
		}
	})

	c.OnResponse(func(r *colly.Response) {
		waiter.Add(1)
		go func() {
			urlHash := getHash(r.Body)
			fPath := filepath.Join(responseLogDir, "response-"+urlHash+".txt")
			r.Save(fPath)
			waiter.Done()
		}()
	})

	c.OnError(func(e *colly.Response, err error) {
		log.Println(strconv.Itoa(e.StatusCode)+" error: ", err)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("\n\tVisiting", r.URL)
		r.Method = "GET"
		r.Headers.Set("User-Agent", "Mozilla/5.0")
	})

	c.Visit(scrpUrl.String())
	waiter.Wait()
}

func saveInfo(fName string, data string) error {
	date := time.Now()
	stringDate := fmt.Sprintf("%d.%d.%d %d-%d-%d ", date.Day(), date.Month(), date.Year(), date.Hour(), date.Minute(), date.Second())
	nFname := "IMGlinks " + stringDate + getHash([]byte(fName))
	fPath := filepath.Join(storageDir, nFname)
	var file *os.File

	if _, err := os.Stat(fPath); err != nil && errors.Is(err, os.ErrNotExist) {
		file, err = os.Create(fPath)
		if err != nil {
			return err
		}
	} else {
		file, err = os.OpenFile(fPath, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	file.WriteString("\n" + data)
	return nil
}

func configureUrl() *url.URL {
	var q url.Values = url.Values{}

	q.Add("q", text)
	q.Add("tbm", "isch")
	q.Add("source", "hp")
	q.Add("hl", "ru")
	q.Add("sclient", "img")

	if timePeriod != "" {
		if timePeriod == "d" || timePeriod == "w" || timePeriod == "m" || timePeriod == "y" {
			q.Add("tbs", "qdr:"+timePeriod)
		} else {
			log.Panic("invalid tbs flag value\nplese enter correct flag\nfor more info enter -? flag")
		}
	}

	urlY, err := url.Parse("https://www.google.ru/search")
	if err != nil {
		log.Panic(err)
	}
	urlY.RawQuery = q.Encode()

	return urlY
}

func getHash(data []byte) string {
	hasher := md5.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}
