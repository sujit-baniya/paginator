package main

import (
	"database/sql"
	"encoding/json"
	"github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/sujit-baniya/paginator"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
	// see https://gorm.io/docs/connecting_to_the_database.html
	dbConn = "host=localhost port=5432 user=postgres dbname=casbin password=postgres"
)

type Campaign struct {
	ID            int64          `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time      `gorm:"primaryKey" json:"created_at"`
	RequestId     string         `gorm:"request_id" json:"request_id"`
	Message       string         `gorm:"message" json:"message" form:"message" query:"message"`
	To            pq.StringArray `gorm:"type:text[]" json:"to" form:"to" query:"to"`
	RecipientType string         `gorm:"recipient_type" json:"recipient_type"`
	From          string         `gorm:"from" json:"from" form:"from" query:"from"`
	NotifyUrl     string         `gorm:"notify_url" json:"notify_url" form:"notify_url" query:"notify_url"`
	// ACTIVE, REJECTED, CANCELED, DONE, SCHEDULED
	Status            string         `gorm:"status" json:"status" form:"status" query:"status"`
	ProviderRate      float32        `gorm:"provider_rate" json:"provider_rate" form:"provider_rate" query:"provider_rate"`
	Rate              float32        `gorm:"rate" json:"rate" form:"rate" query:"rate"`
	ShortenerUrl      string         `json:"shortener_url"`
	ShortenUrls       bool           `json:"shorten_urls,omitempty" form:"shorten_urls"`
	SanitizeContent   bool           `json:"sanitize_content,omitempty" form:"sanitize_content"`
	IsFavourite       bool           `gorm:"is_favourite" json:"is_favourite" form:"is_favourite"`
	UserID            uint           `gorm:"user_id" json:"user_id" form:"user_id" query:"user_id"`
	TotalCount        int            `gorm:"total_count" json:"total_count" form:"total_count" query:"total_count"`
	DeliveredCount    int            `gorm:"delivered_count" json:"delivered_count" form:"delivered_count" query:"delivered_count"`
	FailedCount       int            `gorm:"failed_count" json:"failed_count" form:"failed_count" query:"failed_count"`
	BlockedCount      int            `gorm:"blocked_count" json:"blocked_count" form:"blocked_count" query:"blocked_count"`
	RejectedCount     int            `gorm:"rejected_count" json:"rejected_count" form:"rejected_count" query:"rejected_count"`
	UnsubscribedCount int            `gorm:"unsubscribed_count" json:"unsubscribed_count" form:"unsubscribed_count" query:"unsubscribed_count"`
	CanceledAt        sql.NullTime   `gorm:"canceled_at" json:"canceled_at"`
	SentAt            sql.NullTime   `gorm:"sent_at" json:"sent_at"`
	StartedAt         string         `gorm:"-" json:"started_at"`
	UpdatedAt         time.Time      `gorm:"updated_at" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

func (c *Campaign) Paginate(paging ...paginator.Paging) {

}

func main() {
	var err error
	db, err = gorm.Open(postgres.Open(dbConn), &gorm.Config{})
	http.HandleFunc("/", getBookList)
	err = http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("Error serve: ", err)
	}

}

func getBookList(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		query    = r.URL.Query()
		dbEntity = db
		paging   = paginator.Paging{}
		bookList = struct {
			Data       []*Campaign           `json:"data"`
			Pagination *paginator.Pagination `json:"pagination"`
		}{}
	)

	// get paging params from query
	if len(query.Get("page")) > 0 && query.Get("page") != "" {
		paging.Page, _ = strconv.Atoi(query.Get("page"))
	}
	if len(query.Get("limit")) > 0 && query.Get("limit") != "" {
		paging.Limit, _ = strconv.Atoi(query.Get("limit"))
	}
	if orders, ok := query["order"]; ok || len(orders) > 0 {
		for i := range orders {
			paging.OrderBy = append(paging.OrderBy, orders[i])
		}
	}
	bookList.Pagination, err = paginator.Pages(&paginator.Param{
		DB:     dbEntity,
		Paging: &paging,
	}, &bookList.Data)
	if err != nil {
		log.Fatal("Error get list: ", err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(bookList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}
