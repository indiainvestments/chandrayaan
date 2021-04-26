package corp_news

import (
	"encoding/json"
	"chandrayaan/store"
	strip "github.com/grokify/html-strip-tags-go"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
	"unicode"
)

const endpoint = "https://api.bseindia.com/BseIndiaAPI/api/AnnGetData/w"
const attachmentEndpoint = "https://www.bseindia.com/xml-data/corpfiling/AttachLive"

type BseCorporateNewsFetcherOpts struct {
	HttpClient *http.Client
	Debug      bool
}

type BseCorporateNewsFetcher struct {
	httpClient *http.Client
	debug      bool
}

type BseNewsResponse struct {
	Table []struct {
		AGENDAID         int         `json:"AGENDA_ID"`
		ANNOUNCEMENTTYPE string      `json:"ANNOUNCEMENT_TYPE"`
		ATTACHMENTNAME   string      `json:"ATTACHMENTNAME"`
		CATEGORYNAME     string      `json:"CATEGORYNAME"`
		CRITICALNEWS     int         `json:"CRITICALNEWS"`
		DTTM             string      `json:"DT_TM"`
		DissemDT         string      `json:"DissemDT"`
		FILESTATUS       string      `json:"FILESTATUS"`
		HEADLINE         string      `json:"HEADLINE"`
		MORE             string      `json:"MORE"`
		NEWSID           string      `json:"NEWSID"`
		NEWSSUB          string      `json:"NEWSSUB"`
		NEWSDT           string      `json:"NEWS_DT"`
		NSURL            string      `json:"NSURL"`
		NewsSubmissionDt string      `json:"News_submission_dt"`
		OLD              int         `json:"OLD"`
		PDFFLAG          int         `json:"PDFFLAG"`
		QUARTERID        interface{} `json:"QUARTER_ID"`
		RN               int         `json:"RN"`
		SCRIPCD          int         `json:"SCRIP_CD"`
		SLONGNAME        string      `json:"SLONGNAME"`
		TimeDiff         string      `json:"TimeDiff"`
		TotalPageCnt     int         `json:"TotalPageCnt"`
		XMLNAME          string      `json:"XML_NAME"`
	} `json:"Table"`
}

func (fetcher *BseCorporateNewsFetcher) buildRequest(data store.TickerInfo, opts CorporateNewsOpts) (*http.Request, error) {

	apiUrl, err := url.Parse(endpoint)

	if err != nil {
		return nil, err
	}

	query := apiUrl.Query()
	query.Add("strCat", "-1")
	query.Add("strType", "C")
	query.Add("strSearch", "A")
	query.Add("strPrevDate", opts.From.Format("20060102"))
	query.Add("strToDate", opts.To.Format("20060102"))
	query.Add("strScrip", data.Code)

	apiUrl.RawQuery = query.Encode()

	return http.NewRequest(http.MethodGet, apiUrl.String(), nil)

}

func (fetcher *BseCorporateNewsFetcher) normalize(text string) string {

	text = strings.TrimFunc(text, func(r rune) bool {
		return !unicode.IsGraphic(r)
	})

	text = strip.StripTags(text)

	text = strings.Replace(text, "\u0026", "&", -1)

	return text

}

func (fetcher *BseCorporateNewsFetcher) GetNews(data store.TickerInfo, opts CorporateNewsOpts) (*CorporateNews, error) {
	today := time.Now()
	if opts.To == nil {
		opts.To = &today
	}

	if opts.From == nil {
		fewDaysAgo := today.AddDate(0, 0, -30)
		opts.From = &fewDaysAgo
	}

	apiRequest, err := fetcher.buildRequest(data, opts)
	if err != nil {
		return nil, err
	}

	if fetcher.debug {
		raw, _ := httputil.DumpRequest(apiRequest, true)
		log.Println(string(raw))
	}

	var bseNewsResponse BseNewsResponse

	response, err := fetcher.httpClient.Do(apiRequest)
	if err != nil {
		return nil, err
	}

	if err := json.NewDecoder(response.Body).Decode(&bseNewsResponse); err != nil {
		return nil, err
	}

	var result = CorporateNews{
		Ticker: data.Ticker,
	}

	for _, i := range bseNewsResponse.Table {

		attachment := ""
		if len(i.ATTACHMENTNAME) > 0 {
			attachment = attachmentEndpoint + "/" + i.ATTACHMENTNAME
		}

		result.News = append(result.News, CorporateNewsItem{
			Attachment: attachment,
			Headline:   fetcher.normalize(i.HEADLINE),
			Date:       i.DTTM,
			Category:   i.CATEGORYNAME,
			Id:         i.NEWSID,
			NewsSub:    i.NEWSSUB,
		})
	}

	return &result, nil
}

func NewBseCorporateNewsFetcher(opts ...BseCorporateNewsFetcherOpts) *BseCorporateNewsFetcher {
	var opt BseCorporateNewsFetcherOpts
	if len(opts) >= 1 {
		opt = opts[0]
	}
	if opt.HttpClient == nil {
		opt.HttpClient = &http.Client{Timeout: time.Minute}
	}

	return &BseCorporateNewsFetcher{
		httpClient: opt.HttpClient,
		debug:      opt.Debug,
	}
}
