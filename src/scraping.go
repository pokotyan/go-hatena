package scraping

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	firestore "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/PuerkitoBio/goquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type HatenaLink struct {
	date  string
	users string
	title string
	link  string
	desc  string
}

type FireBase struct {
	opt    option.ClientOption
	ctx    context.Context
	app    *firebase.App
	client *firestore.Client
}

func scraping(date string) HatenaLink {
	doc, err := goquery.NewDocument("http://b.hatena.ne.jp/hotentry/it/" + date)
	if err != nil {
		fmt.Print("url scarapping failed")
	}
	users := doc.Find("#container > div.wrapper > div > div.entrylist-main > section > div > ul > li > div > div.entrylist-contents-main > span > a > span").Text()
	selection := doc.Find("#container > div.wrapper > div > div.entrylist-main > section > div > ul > li > div > div.entrylist-contents-main > h3 > a")
	title := selection.Text()
	link, _ := selection.Attr("href")
	desc := doc.Find("#container > div.wrapper > div > div.entrylist-main > section > div > ul > li > div > div.entrylist-contents-main > div.entrylist-contents-body > a > p.entrylist-contents-description").Text()
	data := HatenaLink{date, users, title, link, desc}

	return data
}

func initFireBase(keyFileJSON []byte) *FireBase {
	opt := option.WithCredentialsJSON(keyFileJSON)
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)

	if err != nil {
		fmt.Printf("error initializing app: %v", err)
	}

	client, err := app.Firestore(ctx)

	if err != nil {
		log.Fatalln(err)
	}

	return &FireBase{opt, ctx, app, client}
}

func (f *FireBase) findOne(date string) (interface{}, error) {
	iter := f.client.Collection("hatena_link").Where("date", "==", date).Documents(f.ctx)
	var result interface{}

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		result = doc.Data()
	}

	if result == nil {
		return nil, nil
	}

	return result, nil
}

func (f *FireBase) create(hl HatenaLink) {
	f.client.Collection("hatena_link").Add(f.ctx, map[string]string{
		"date":        hl.date,
		"users_count": hl.users,
		"title":       hl.title,
		"link":        hl.link,
		"desc":        hl.desc,
	})
}

func Run() {
	sEnc := os.Getenv("FIREBASE_KEYFILE_JSON")
	sDec, _ := base64.StdEncoding.DecodeString(sEnc)

	f := initFireBase(sDec)

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	formatedDate := yesterday.Format("20060102")
	data, err := f.findOne(formatedDate)

	if err != nil {
		fmt.Printf("データ取得に失敗 err: %s", err)
	}

	if data != nil {
		fmt.Printf("%vのデータはすでに保存ずみ\n", formatedDate)
	}

	if data == nil {
		fmt.Printf("%vのデータをスクレイピング、保存します\n", formatedDate)
		hl := scraping(formatedDate)
		f.create(hl)
	}

	defer f.client.Close()
}
