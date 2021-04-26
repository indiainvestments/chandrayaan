package corp_action

import (
	"encoding/json"
	"chandrayaan/store"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type BseCorporateActionFetcherOpts struct {
	HttpClient *http.Client
	Debug      bool
}

type BseCorporateActionFetcher struct {
	httpClient *http.Client
	debug      bool
}

const endpoint = "https://api.bseindia.com/BseIndiaAPI/api/CorporateAction/w"

type BseCorporateActionResponse struct {
	Table []struct {
		PurposeName string  `json:"purpose_name"`
		BCRDFrom    string  `json:"BCRD_from"`
		Amount      float64 `json:"Amount"`
	} `json:"Table"`
	Table1 []struct {
		XTYPE    string `json:"XTYPE"`
		BCRDFROM string `json:"BCRD_FROM"`
		VALUE    string `json:"VALUE"`
	} `json:"Table1"`
	Table2 []struct {
		ScripCode   int    `json:"scrip_code"`
		SLongName   string `json:"sLongName"`
		BCRD        string `json:"BCRD"`
		PurposeCode string `json:"purpose_code"`
		ShortName   string `json:"short_name"`
		ExDate      string `json:"Ex_date"`
		Purpose     string `json:"purpose"`
		Details     string `json:"Details"`
		PAYMENTDATE string `json:"PAYMENT_DATE"`
	} `json:"Table2"`
}

func (fetcher *BseCorporateActionFetcher) buildRequest(data store.TickerInfo) (*http.Request, error) {

	apiUrl, err := url.Parse(endpoint)

	if err != nil {
		return nil, err
	}

	query := apiUrl.Query()
	query.Add("scripcode", data.Code)
	apiUrl.RawQuery = query.Encode()

	return http.NewRequest(http.MethodGet, apiUrl.String(), nil)

}

func (fetcher *BseCorporateActionFetcher) GetNews(data store.TickerInfo) (*CorporateAction, error) {

	apiRequest, err := fetcher.buildRequest(data)
	if err != nil {
		return nil, err
	}

	if fetcher.debug {
		raw, _ := httputil.DumpRequest(apiRequest, true)
		log.Println(string(raw))
	}

	var bseNewsResponse BseCorporateActionResponse

	response, err := fetcher.httpClient.Do(apiRequest)
	if err != nil {
		return nil, err
	}

	if err := json.NewDecoder(response.Body).Decode(&bseNewsResponse); err != nil {
		return nil, err
	}

	var result = CorporateAction{
		Ticker: data.Ticker,
	}

	for _, i := range bseNewsResponse.Table2 {
		result.Actions = append(result.Actions, CorporateActionItem{
			ExDate:      i.ExDate,
			Purpose:     i.Purpose,
			Details:     i.Details,
			PaymentDate: i.PAYMENTDATE,
		})
	}

	return &result, nil
}

func NewBseCorporateActionFetcher(opts ...BseCorporateActionFetcherOpts) *BseCorporateActionFetcher {
	var opt BseCorporateActionFetcherOpts
	if len(opts) >= 1 {
		opt = opts[0]
	}
	if opt.HttpClient == nil {
		opt.HttpClient = &http.Client{Timeout: time.Minute}
	}

	return &BseCorporateActionFetcher{
		httpClient: opt.HttpClient,
		debug:      opt.Debug,
	}
}
